package thing

import (
	"fmt"

	"github.com/arduino/iot-cloud-cli/command/thing"
	"github.com/spf13/cobra"
)

var bindFlags struct {
	id       string
	deviceID string
}

func initBindCommand() *cobra.Command {
	bindCommand := &cobra.Command{
		Use:   "bind",
		Short: "Bind a thing to a device",
		Long:  "Bind a thing to a device on Arduino IoT Cloud",
		RunE:  runBindCommand,
	}
	bindCommand.Flags().StringVarP(&bindFlags.id, "id", "i", "", "Thing ID")
	bindCommand.Flags().StringVarP(&bindFlags.deviceID, "device-id", "d", "", "Device ID")
	bindCommand.MarkFlagRequired("id")
	bindCommand.MarkFlagRequired("device-id")
	return bindCommand
}

func runBindCommand(cmd *cobra.Command, args []string) error {
	fmt.Printf("Binding thing %s to device%s\n", bindFlags.id, bindFlags.deviceID)

	params := &thing.BindParams{
		ID:       bindFlags.id,
		DeviceID: bindFlags.deviceID,
	}
	err := thing.Bind(params)
	if err != nil {
		return err
	}

	fmt.Println("Thing-Device bound successfully updated")
	return nil
}
