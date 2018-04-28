package motor

import (
	"fmt"
	"os"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

// https://github.com/EliasKotlyar/Xiaomi-Dafang-Software/blob/master/motor/motor.c
type Command uint32
type Direction uint32
type Axis uint32

type Status struct {
	YMax   int32
	YMin   int32
	XMax   int32
	XMin   int32
	XSteps int32
	YSteps int32
}

const (
	StopCommand Command = 1 + iota
	ResetCommand
	MoveCommand
	GetStatusCommand
	SpeedCommand
)
const (
	UpDirection Direction = iota
	DownDirection
	LeftDirection
	RightDirection
)

const (
	HorizontalAxis Axis = 1 + iota
	VerticalAxis
)

const InitSpeed = 1000

type ControllerAxis struct {
	max int32
}

func (c *ControllerAxis) Max() int32 {
	return c.max
}

type Controller struct {
	movement    sync.Mutex
	calibration sync.Mutex
	fd          int
	speed       uint32
	calibrated  time.Time
	XAxis       ControllerAxis
	YAxis       ControllerAxis
}

type Movement struct {
	Direction Direction
	Steps     int32
	Speed     uint32
}

func NewController() (*Controller, error) {
	fd, err := unix.Open("/dev/motor", os.O_RDONLY|unix.O_CLOEXEC, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open motor control: %v", err)
	}
	return &Controller{
		fd:    fd,
		speed: 1000,
	}, nil
}

func (controller *Controller) goTo(x, y int32) error {
	status, err := controller.Status()
	if err != nil {
		return err
	}
	err = controller.goToAxis(x, status.XSteps, controller.Right, controller.Left)
	if err != nil {
		return err
	}
	return controller.goToAxis(y, status.YSteps, controller.Up, controller.Down)
}

func (controller *Controller) goToAxis(target int32, current int32, incFunc, decFunc func(int32) error) error {

	steps := target - current
	switch {
	case steps == 0:
		return nil
	case steps > 0:
		return incFunc(steps)
	case steps < 0:
		return decFunc(steps)
	}
	return fmt.Errorf("unable to decide how to walk %s steps", steps)
}

func (controller *Controller) Center() error {
	return controller.goTo(controller.XAxis.max/2, controller.YAxis.max/2)
}
func (controller *Controller) Stop() error {
	return controller.sendCommand(StopCommand, unsafe.Pointer(nil))
}
func (controller *Controller) Reset() error {
	return controller.sendCommand(ResetCommand, unsafe.Pointer(nil))
}
func (controller *Controller) Up(steps int32) error {
	controller.movement.Lock()
	defer controller.movement.Unlock()
	mov := Movement{
		Direction: UpDirection,
		Speed:     controller.speed,
		Steps:     steps,
	}
	err := controller.sendCommand(MoveCommand, unsafe.Pointer(&mov))
	if err != nil {
		return err
	}
	return controller.wait()
}

func (controller *Controller) Down(steps int32) error {
	controller.movement.Lock()
	defer controller.movement.Unlock()

	mov := Movement{
		Direction: DownDirection,
		Speed:     controller.speed,
		Steps:     steps,
	}
	err := controller.sendCommand(MoveCommand, unsafe.Pointer(&mov))
	if err != nil {
		return err
	}
	return controller.wait()
}

func (controller *Controller) Right(steps int32) error {
	controller.movement.Lock()
	defer controller.movement.Unlock()

	mov := Movement{
		Direction: RightDirection,
		Speed:     controller.speed,
		Steps:     steps,
	}
	err := controller.sendCommand(MoveCommand, unsafe.Pointer(&mov))
	if err != nil {
		return err
	}
	return controller.wait()
}

func (controller *Controller) Left(steps int32) error {
	controller.movement.Lock()
	defer controller.movement.Unlock()

	mov := Movement{
		Direction: LeftDirection,
		Speed:     controller.speed,
		Steps:     steps,
	}
	err := controller.sendCommand(MoveCommand, unsafe.Pointer(&mov))
	if err != nil {
		return err
	}
	return controller.wait()
}
func (controller *Controller) Status() (Status, error) {
	controller.calibration.Lock()
	defer controller.calibration.Unlock()
	status := Status{}
	err := controller.sendCommand(GetStatusCommand, unsafe.Pointer(&status))
	if err != nil {
		return status, err
	}
	return status, nil
}
func (controller *Controller) SetX(target int32) error {
	status, err := controller.Status()
	if err != nil {
		return err
	}
	return controller.goToAxis(target, status.XSteps, controller.Right, controller.Left)

}
func (controller *Controller) SetY(target int32) error {
	status, err := controller.Status()
	if err != nil {
		return err
	}
	return controller.goToAxis(target, status.YSteps, controller.Up, controller.Down)
}

func (controller *Controller) wait() error {
	status, err := controller.Status()
	if err != nil {
		return err
	}
	for range time.Tick(200 * time.Millisecond) {
		current, err := controller.Status()
		if err != nil {
			return err
		}
		if current.YSteps == status.YSteps && current.XSteps == status.XSteps {
			return nil
		}
		status = current
	}
	return nil
}
func (controller *Controller) Speed(speed int32) error {
	_speed := speed
	return controller.sendCommand(SpeedCommand, unsafe.Pointer(&_speed))
}
func (controller *Controller) Calibrate() error {
	controller.calibration.Lock()
	defer controller.calibration.Unlock()
	controller.Stop()
	err := controller.calibrateY()
	if err != nil {
		return err
	}
	err = controller.calibrateX()
	if err != nil {
		return err
	}
	controller.calibrated = time.Now()
	return nil
}
func (controller *Controller) calibrateY() error {
	var status Status
	var err error
	controller.Down(5000)
	for range time.Tick(1 * time.Second) {
		status, err = controller.Status()
		if err != nil {
			return err
		}
		if status.YMin == 1 {
			break
		}
	}
	controller.Reset()
	controller.Up(5000)
	for range time.Tick(1 * time.Second) {
		status, err = controller.Status()
		if err != nil {
			return err
		}
		if status.YMax == 1 {
			break
		}
	}
	max := status.YSteps
	controller.Down(5000)
	for range time.Tick(1 * time.Second) {
		status, err = controller.Status()
		if err != nil {
			return err
		}
		if status.YMin == 1 {
			break
		}

	}
	controller.YAxis.max = max
	return nil
}

func (controller *Controller) calibrateX() error {
	var status Status
	var err error
	controller.Left(5000)
	for range time.Tick(1 * time.Second) {
		status, err = controller.Status()
		if err != nil {
			return err
		}
		if status.XMin == 1 {
			break
		}
	}
	controller.Reset()
	controller.Right(5000)
	for range time.Tick(1 * time.Second) {
		status, err = controller.Status()
		if err != nil {
			return err
		}
		if status.XMax == 1 {
			break
		}
	}
	controller.XAxis.max = status.XSteps
	controller.Left(5000)
	for range time.Tick(1 * time.Second) {
		status, err = controller.Status()
		if err != nil {
			return err
		}
		if status.XMin == 1 {
			break
		}
	}
	return nil
}

func (controller *Controller) sendCommand(cmd Command, payload unsafe.Pointer) error {
	_, _, errorp := unix.Syscall(unix.SYS_IOCTL,
		uintptr(controller.fd),
		uintptr(int(cmd)),
		uintptr(payload))

	if errorp.Error() != "errno 0" {
		return fmt.Errorf("ioctl returned: %v", errorp)
	}
	return nil
}
