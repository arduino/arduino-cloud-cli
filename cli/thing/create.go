package thing

import (
	"fmt"

	"github.com/arduino/iot-cloud-cli/command/thing"
	"github.com/spf13/cobra"
)

var createFlags struct {
	name     string
	device   string
	template string
	clone    string
}

func initCreateCommand() *cobra.Command {
	createCommand := &cobra.Command{
		Use:   "create",
		Short: "Create a thing",
		Long:  "Create a thing for Arduino IoT Cloud",
		RunE:  runCreateCommand,
	}
	createCommand.Flags().StringVarP(&createFlags.name, "name", "n", "", "Thing name")
	createCommand.Flags().StringVarP(&createFlags.device, "device", "d", "", "ID of Device to bind to the new thing")
	createCommand.Flags().StringVarP(&createFlags.clone, "clone", "c", "", "ID of Thing to be cloned")
	createCommand.Flags().StringVarP(&createFlags.template, "template", "t", "", "File containing a thing template")
	createCommand.MarkFlagRequired("name")
	return createCommand
}

func runCreateCommand(cmd *cobra.Command, args []string) error {
	fmt.Printf("Creating thing with name %s\n", createFlags.name)

	params := &thing.CreateParams{
		Name:     createFlags.name,
		Device:   createFlags.device,
		Template: createFlags.template,
		Clone:    createFlags.clone,
	}

	thingID, err := thing.Create(params)
	if err != nil {
		return err
	}

	fmt.Printf("IoT Cloud thing created with ID: %s\n", thingID)
	return nil
}
