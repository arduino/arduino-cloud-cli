package device

import (
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/iot-cloud-cli/command/device"
	"github.com/sirupsen/logrus"
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
		Run:   runDeleteCommand,
	}
	deleteCommand.Flags().StringVarP(&deleteFlags.id, "id", "i", "", "Device ID")
	deleteCommand.MarkFlagRequired("id")
	return deleteCommand
}

func runDeleteCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Deleting device %s\n", deleteFlags.id)

	params := &device.DeleteParams{ID: deleteFlags.id}
	err := device.Delete(params)
	if err != nil {
		feedback.Errorf("Error during device delete: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Info("Device successfully deleted")
}
