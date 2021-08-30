package thing

import (
	"fmt"

	"github.com/arduino/iot-cloud-cli/command/thing"
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
		RunE:  runCloneCommand,
	}
	cloneCommand.Flags().StringVarP(&cloneFlags.name, "name", "n", "", "Thing name")
	cloneCommand.Flags().StringVarP(&cloneFlags.cloneID, "clone-id", "c", "", "ID of Thing to be cloned")
	cloneCommand.MarkFlagRequired("name")
	cloneCommand.MarkFlagRequired("clone-id")
	return cloneCommand
}

func runCloneCommand(cmd *cobra.Command, args []string) error {
	fmt.Printf("Cloning thing %s into a new thing called %s\n", cloneFlags.cloneID, cloneFlags.name)

	params := &thing.CloneParams{
		Name:    cloneFlags.name,
		CloneID: cloneFlags.cloneID,
	}

	thingID, err := thing.Clone(params)
	if err != nil {
		return err
	}

	fmt.Printf("IoT Cloud thing created with ID: %s\n", thingID)
	return nil
}
