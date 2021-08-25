package thing

import (
	"fmt"
	"strings"

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/iot-cloud-cli/command/thing"
	"github.com/spf13/cobra"
)

var listFlags struct {
	ids        []string
	deviceID   string
	properties bool
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
	// list only the things associated to the passed device id
	listCommand.Flags().StringVarP(&listFlags.deviceID, "device", "d", "", "ID of Device associated to the thing to be retrieved")
	listCommand.Flags().BoolVarP(&listFlags.properties, "properties", "p", false, "Show thing properties")
	return listCommand
}

func runListCommand(cmd *cobra.Command, args []string) error {
	fmt.Println("Listing things")

	params := &thing.ListParams{
		IDs:        listFlags.ids,
		Properties: listFlags.properties,
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
	if listFlags.properties {
		h = append(h, "Properties")
	}
	t.SetHeader(h...)

	for _, thing := range r.things {
		r := []interface{}{thing.Name, thing.ID, thing.DeviceID}
		if listFlags.properties {
			r = append(r, strings.Join(thing.Properties, ", "))
		}
		t.AddRow(r...)
	}
	return t.Render()
}
