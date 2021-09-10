package thing

import (
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/iot-cloud-cli/command/thing"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var deleteFlags struct {
	id string
}

func initDeleteCommand() *cobra.Command {
	deleteCommand := &cobra.Command{
		Use:   "delete",
		Short: "Delete a thing",
		Long:  "Delete a thing from Arduino IoT Cloud",
		Run:   runDeleteCommand,
	}
	deleteCommand.Flags().StringVarP(&deleteFlags.id, "id", "i", "", "Thing ID")
	deleteCommand.MarkFlagRequired("id")
	return deleteCommand
}

func runDeleteCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Deleting thing %s\n", deleteFlags.id)

	params := &thing.DeleteParams{ID: deleteFlags.id}
	err := thing.Delete(params)
	if err != nil {
		feedback.Errorf("Error during thing delete: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Info("Thing successfully deleted")
}
