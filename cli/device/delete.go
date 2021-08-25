package device

import (
	"fmt"

	"github.com/arduino/iot-cloud-cli/command/device"
	"github.com/spf13/cobra"
)

var deleteFlags struct {
	id string
}

func initDeleteCommand() *cobra.Command {
	deleteCommand := &cobra.Command{
		Use:   "delete",
		Short: "Delete a device",
		Long:  "Delete a device from Arduino IoT Cloud",
		RunE:  runDeleteCommand,
	}
	deleteCommand.Flags().StringVarP(&deleteFlags.id, "id", "i", "", "Device ID")
	deleteCommand.MarkFlagRequired("id")
	return deleteCommand
}

func runDeleteCommand(cmd *cobra.Command, args []string) error {
	fmt.Printf("Deleting device %s\n", deleteFlags.id)

	params := &device.DeleteParams{ID: deleteFlags.id}
	err := device.Delete(params)
	if err != nil {
		return err
	}

	fmt.Println("Device successfully deleted")
	return nil
}
