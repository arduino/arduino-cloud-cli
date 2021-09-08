package thing

import (
	"fmt"

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/iot-cloud-cli/command/thing"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var createFlags struct {
	name     string
	template string
}

func initCreateCommand() *cobra.Command {
	createCommand := &cobra.Command{
		Use:   "create",
		Short: "Create a thing from a template",
		Long:  "Create a thing from a template for Arduino IoT Cloud",
		RunE:  runCreateCommand,
	}
	createCommand.Flags().StringVarP(&createFlags.name, "name", "n", "", "Thing name")
	createCommand.Flags().StringVarP(
		&createFlags.template,
		"template",
		"t",
		"",
		"File containing a thing template, JSON and YAML format are supported",
	)
	createCommand.MarkFlagRequired("template")
	return createCommand
}

func runCreateCommand(cmd *cobra.Command, args []string) error {
	logrus.Infof("Creating thing from template %s\n", createFlags.template)

	params := &thing.CreateParams{
		Template: createFlags.template,
	}
	if createFlags.name != "" {
		params.Name = &createFlags.name
	}

	thing, err := thing.Create(params)
	if err != nil {
		return err
	}

	logrus.Infof("IoT Cloud thing created with ID: %s\n", thing.ID)
	feedback.PrintResult(createResult{thing})
	return nil
}

type createResult struct {
	thing *thing.ThingInfo
}

func (r createResult) Data() interface{} {
	return r.thing
}

func (r createResult) String() string {
	return fmt.Sprintf("IoT Cloud thing created with ID: %s", r.thing.ID)
}
