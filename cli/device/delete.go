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
	ids []string
}

func initDeleteCommand() *cobra.Command {
	deleteCommand := &cobra.Command{
		Use:   "delete",
		Short: "Delete devices",
		Long:  "Delete a list of devices from Arduino IoT Cloud",
		Run:   runDeleteCommand,
	}
	deleteCommand.Flags().StringSliceVarP(&deleteFlags.ids, "ids", "i", []string{}, "List of comma-separated device IDs to be deleted")
	deleteCommand.MarkFlagRequired("ids")
	return deleteCommand
}

func runDeleteCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Deleting devices %v\n", deleteFlags.ids)

	params := &device.DeleteParams{IDs: deleteFlags.ids}
	err := device.Delete(params)
	if err != nil {
		feedback.Errorf("Error during device delete: %s", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Info("Devices successfully deleted")
}
