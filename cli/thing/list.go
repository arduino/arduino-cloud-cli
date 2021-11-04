// This file is part of arduino-cloud-cli.
//
// Copyright (C) 2021 ARDUINO SA (http://www.arduino.cc/)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package thing

import (
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/arduino-cloud-cli/command/thing"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var listFlags struct {
	ids       []string
	deviceID  string
	variables bool
	tags      map[string]string
}

func initListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List things",
		Long:  "List things on Arduino IoT Cloud",
		Run:   runListCommand,
	}
	// list only the things corresponding to the passed ids
	listCommand.Flags().StringSliceVarP(&listFlags.ids, "ids", "i", []string{}, "List of thing IDs to be retrieved")
	// list only the thing associated to the passed device id
	listCommand.Flags().StringVarP(&listFlags.deviceID, "device-id", "d", "", "ID of Device associated to the thing to be retrieved")
	listCommand.Flags().BoolVarP(&listFlags.variables, "show-variables", "s", false, "Show thing variables")
	// list only the things that have all the passed tags
	listCommand.Flags().StringToStringVar(
		&listFlags.tags,
		"tags",
		nil,
		"List of comma-separated tags. A tag has this format: <key>=<value>",
	)
	return listCommand
}

func runListCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Listing things")

	params := &thing.ListParams{
		IDs:       listFlags.ids,
		Variables: listFlags.variables,
		Tags:      listFlags.tags,
	}
	if listFlags.deviceID != "" {
		params.DeviceID = &listFlags.deviceID
	}

	things, err := thing.List(params)
	if err != nil {
		feedback.Errorf("Error during thing list: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.PrintResult(result{things})
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
