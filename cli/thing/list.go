package thing

import (
	"strings"

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/iot-cloud-cli/command/thing"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var listFlags struct {
	ids       []string
	deviceID  string
	variables bool
}

func initListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List things",
		Long:  "List things on Arduino IoT Cloud",
		RunE:  runListCommand,
	}
	// list only the things corresponding to the passed ids
	listCommand.Flags().StringSliceVarP(&listFlags.ids, "ids", "i", []string{}, "List of thing IDs to be retrieved")
	// list only the thing associated to the passed device id
	listCommand.Flags().StringVarP(&listFlags.deviceID, "device-id", "d", "", "ID of Device associated to the thing to be retrieved")
	listCommand.Flags().BoolVarP(&listFlags.variables, "show-variables", "s", false, "Show thing variables")
	return listCommand
}

func runListCommand(cmd *cobra.Command, args []string) error {
	logrus.Info("Listing things")

	params := &thing.ListParams{
		IDs:       listFlags.ids,
		Variables: listFlags.variables,
	}
	if listFlags.deviceID != "" {
		params.DeviceID = &listFlags.deviceID
	}

	things, err := thing.List(params)
	if err != nil {
		return err
	}

	feedback.PrintResult(result{things})
	return nil
}

type result struct {
	things []thing.ThingInfo
}

func (r result) Data() interface{} {
	return r.things
}

func (r result) String() string {
	if len(r.things) == 0 {
		return "No things found."
	}
	t := table.New()

	h := []interface{}{"Name", "ID", "Device"}
	if listFlags.variables {
		h = append(h, "Variables")
	}
	t.SetHeader(h...)

	for _, thing := range r.things {
		r := []interface{}{thing.Name, thing.ID, thing.DeviceID}
		if listFlags.variables {
			r = append(r, strings.Join(thing.Variables, ", "))
		}
		t.AddRow(r...)
	}
	return t.Render()
}
