package main

import (
	"log"
	"os"

	"text/template"

	"github.com/jbonachera/dafang/motor"
	"github.com/spf13/cobra"
)

func main() {
	c, err := motor.NewController()
	if err != nil {
		log.Fatalf("failed to initialize motor: %v", err)
	}
	rootCmd := &cobra.Command{
		Use:   "motor",
		Short: "Control the camera's motor",
	}

	upCmd := &cobra.Command{
		Use:   "up",
		Short: "move up",
		Run: func(cmd *cobra.Command, args []string) {
			steps, _ := cmd.Flags().GetInt32("steps")
			c.Up(steps)
		},
	}
	downCmd := &cobra.Command{
		Use:   "down",
		Short: "move down",
		Run: func(cmd *cobra.Command, args []string) {
			steps, _ := cmd.Flags().GetInt32("steps")
			c.Down(steps)
		},
	}
	leftCmd := &cobra.Command{
		Use:   "left",
		Short: "move left",
		Run: func(cmd *cobra.Command, args []string) {
			steps, _ := cmd.Flags().GetInt32("steps")
			c.Left(steps)
		},
	}
	rightCmd := &cobra.Command{
		Use:   "right",
		Short: "move right",
		Run: func(cmd *cobra.Command, args []string) {
			steps, _ := cmd.Flags().GetInt32("steps")
			c.Right(steps)
		},
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "display motor status",
		Run: func(cmd *cobra.Command, args []string) {
			status, err := c.Status()
			if err == nil {
				tmp, err := template.New("").Parse(`YMax  : {{ .YMax }}
YMin  : {{ .YMin }}
XMax  : {{ .XMax }}
XMin  : {{ .XMin }}
XSteps: {{ .XSteps }}
YSteps: {{ .YSteps }}
`)
				if err != nil {
					log.Fatal(err)
				}
				tmp.Execute(os.Stdout, status)
			}
		},
	}
	calibrateCmd := &cobra.Command{
		Use:   "calibrate",
		Short: "calibrate the motor state",
		Run: func(cmd *cobra.Command, args []string) {
			err := c.Calibrate()
			if err != nil {
				log.Fatal(err)
			}
			err = c.Center()
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "reset the motor state",
		Run: func(cmd *cobra.Command, args []string) {
			c.Reset()
		},
	}
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "stop the motor movment",
		Run: func(cmd *cobra.Command, args []string) {
			c.Stop()
		},
	}
	speedCmd := &cobra.Command{
		Use:   "speed",
		Short: "set the motor movment speed",
		Run: func(cmd *cobra.Command, args []string) {
			speed, _ := cmd.Flags().GetInt32("speed")
			c.Speed(speed)
		},
	}
	speedCmd.Flags().Int32("speed", 1200, "motor speed")

	rootCmd.AddCommand(upCmd)
	rootCmd.AddCommand(downCmd)
	rootCmd.AddCommand(leftCmd)
	rootCmd.AddCommand(rightCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(calibrateCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(speedCmd)
	rootCmd.PersistentFlags().Int32P("steps", "s", 100, "steps to walk")
	rootCmd.Execute()
}
