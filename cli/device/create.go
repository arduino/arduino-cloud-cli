package device

import (
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/iot-cloud-cli/command/device"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var createFlags struct {
	port string
	name string
	fqbn string
}

func initCreateCommand() *cobra.Command {
	createCommand := &cobra.Command{
		Use:   "create",
		Short: "Create a device",
		Long:  "Create a device for Arduino IoT Cloud",
		Run:   runCreateCommand,
	}
	createCommand.Flags().StringVarP(&createFlags.port, "port", "p", "", "Device port")
	createCommand.Flags().StringVarP(&createFlags.name, "name", "n", "", "Device name")
	createCommand.Flags().StringVarP(&createFlags.fqbn, "fqbn", "b", "", "Device fqbn")
	createCommand.MarkFlagRequired("name")
	return createCommand
}

func runCreateCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Creating device with name %s\n", createFlags.name)

	params := &device.CreateParams{
		Name: createFlags.name,
	}
	if createFlags.port != "" {
		params.Port = &createFlags.port
	}
	if createFlags.fqbn != "" {
		params.Fqbn = &createFlags.fqbn
	}

	dev, err := device.Create(params)
	if err != nil {
		feedback.Errorf("Error during device create: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.PrintResult(createResult{dev})
}

type createResult struct {
	device *device.DeviceInfo
}

func (r createResult) Data() interface{} {
	return r.device
}

func (r createResult) String() string {
	return fmt.Sprintf("IoT Cloud device created with ID: %s", r.device.ID)
}
