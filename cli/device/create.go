package device

import (
	"fmt"

	"github.com/arduino/iot-cloud-cli/command/device"
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
		RunE:  runCreateCommand,
	}
	createCommand.Flags().StringVarP(&createFlags.port, "port", "p", "", "Device port")
	createCommand.Flags().StringVarP(&createFlags.name, "name", "n", "", "Device name")
	createCommand.Flags().StringVarP(&createFlags.fqbn, "fqbn", "b", "", "Device fqbn")
	createCommand.MarkFlagRequired("name")
	return createCommand
}

func runCreateCommand(cmd *cobra.Command, args []string) error {
	fmt.Printf("Creating device with name %s\n", createFlags.name)

	params := &device.CreateParams{
		Name: createFlags.name,
	}
	if createFlags.port != "" {
		params.Port = &createFlags.port
	}
	if createFlags.fqbn != "" {
		params.Fqbn = &createFlags.fqbn
	}

	devID, err := device.Create(params)
	if err != nil {
		return err
	}

	fmt.Printf("IoT Cloud device created with ID: %s\n", devID)
	return nil
}
