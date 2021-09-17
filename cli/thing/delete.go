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
	ids []string
}

func initDeleteCommand() *cobra.Command {
	deleteCommand := &cobra.Command{
		Use:   "delete",
		Short: "Delete things",
		Long:  "Delete a list of things from Arduino IoT Cloud",
		Run:   runDeleteCommand,
	}
	deleteCommand.Flags().StringSliceVarP(&deleteFlags.ids, "ids", "i", []string{}, "List of comma-separated thing IDs to be deleted")
	deleteCommand.MarkFlagRequired("ids")
	return deleteCommand
}

func runDeleteCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Deleting things %v\n", deleteFlags.ids)

	params := &thing.DeleteParams{IDs: deleteFlags.ids}
	err := thing.Delete(params)
	if err != nil {
		feedback.Errorf("Error during thing delete: %s", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Info("Things successfully deleted")
}
