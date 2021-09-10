package thing

import (
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/iot-cloud-cli/command/thing"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cloneFlags struct {
	name    string
	cloneID string
}

func initCloneCommand() *cobra.Command {
	cloneCommand := &cobra.Command{
		Use:   "clone",
		Short: "Clone a thing",
		Long:  "Clone a thing for Arduino IoT Cloud",
		Run:   runCloneCommand,
	}
	cloneCommand.Flags().StringVarP(&cloneFlags.name, "name", "n", "", "Thing name")
	cloneCommand.Flags().StringVarP(&cloneFlags.cloneID, "clone-id", "c", "", "ID of Thing to be cloned")
	cloneCommand.MarkFlagRequired("name")
	cloneCommand.MarkFlagRequired("clone-id")
	return cloneCommand
}

func runCloneCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Cloning thing %s into a new thing called %s\n", cloneFlags.cloneID, cloneFlags.name)

	params := &thing.CloneParams{
		Name:    cloneFlags.name,
		CloneID: cloneFlags.cloneID,
	}

	thing, err := thing.Clone(params)
	if err != nil {
		feedback.Errorf("Error during thing clone: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.PrintResult(cloneResult{thing})
}

type cloneResult struct {
	thing *thing.ThingInfo
}

func (r cloneResult) Data() interface{} {
	return r.thing
}

func (r cloneResult) String() string {
	return fmt.Sprintf("IoT Cloud thing created with ID: %s", r.thing.ID)
}
