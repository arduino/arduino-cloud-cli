package thing

import (
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/thing"
	"github.com/sirupsen/logrus"
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
		Run:   runBindCommand,
	}
	bindCommand.Flags().StringVarP(&bindFlags.id, "id", "i", "", "Thing ID")
	bindCommand.Flags().StringVarP(&bindFlags.deviceID, "device-id", "d", "", "Device ID")
	bindCommand.MarkFlagRequired("id")
	bindCommand.MarkFlagRequired("device-id")
	return bindCommand
}

func runBindCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Binding thing %s to device %s\n", bindFlags.id, bindFlags.deviceID)

	params := &thing.BindParams{
		ID:       bindFlags.id,
		DeviceID: bindFlags.deviceID,
	}
	err := thing.Bind(params)
	if err != nil {
		feedback.Errorf("Error during thing bind: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Info("Thing-Device bound successfully updated")
}
