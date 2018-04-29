package daylight

import (
	"errors"
	"io"
	"log"
	"os"
	"time"
)

const (
	DAYLIGHT_SENSOR = "/dev/jz_adc_aux_0"
)

type Reporter interface {
	Percent() int
	Raw() []byte
	Stop()
}

type reporter struct {
	device io.ReadCloser
	cancel chan struct{}
	done   chan struct{}
	value  []byte
}

func NewReporter() (Reporter, error) {
	fd, err := os.Open(DAYLIGHT_SENSOR)
	if err != nil {
		return nil, errors.New("failed to open light sensor at " + DAYLIGHT_SENSOR)
	}
	r := &reporter{
		cancel: make(chan struct{}),
		done:   make(chan struct{}),
		device: fd,
	}
	r.start()
	return r, nil
}

func (r *reporter) start() {
	ticker := time.Tick(500 * time.Millisecond)
	go func() {
		defer func() {
			r.device.Close()
			close(r.done)
		}()
		buf := make([]byte, 2)
		for {
			select {
			case <-ticker:
				_, err := r.device.Read(buf)
				if err != nil {
					log.Println(err)
				}
				r.value = buf
			case <-r.cancel:
				return
			}
		}
	}()
}

func (r *reporter) Stop() {
	close(r.cancel)
	<-r.done
}
func (r *reporter) Raw() []byte {
	return r.value
}
func (r *reporter) Percent() int {
	var value int
	value = int(r.value[1])
	value = value << 8
	value += int(r.value[0])
	percent := (value * 100) / 4096
	return percent
}
