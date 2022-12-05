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
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/arduino-cloud-cli/command/thing"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type listFlags struct {
	ids       []string
	deviceID  string
	variables bool
	tags      map[string]string
}

func initListCommand() *cobra.Command {
	flags := &listFlags{}
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List things",
		Long:  "List things on Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runListCommand(flags); err != nil {
				feedback.Errorf("Error during thing list: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	// list only the things corresponding to the passed ids
	listCommand.Flags().StringSliceVarP(&flags.ids, "ids", "i", nil, "List of thing IDs to be retrieved")
	// list only the thing associated to the passed device id
	listCommand.Flags().StringVarP(&flags.deviceID, "device-id", "d", "", "ID of Device associated to the thing to be retrieved")
	listCommand.Flags().BoolVarP(&flags.variables, "show-variables", "s", false, "Show thing variables")
	listCommand.Flags().StringToStringVar(
		&flags.tags,
		"tags",
		nil,
		"Comma-separated list of tags with format <key>=<value>.\n"+
			"List only things that match the provided tags.",
	)
	return listCommand
}

func runListCommand(flags *listFlags) error {
	logrus.Info("Listing things")

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	params := &thing.ListParams{
		IDs:       flags.ids,
		Variables: flags.variables,
		Tags:      flags.tags,
	}
	if flags.deviceID != "" {
		params.DeviceID = &flags.deviceID
	}

	things, err := thing.List(context.TODO(), params, cred)
	if err != nil {
		return err
	}

	feedback.PrintResult(result{things: things, variables: flags.variables})
	return nil
}

type result struct {
	things    []thing.ThingInfo
	variables bool
}

func (r result) Data() interface{} {
	return r.things
}

func (r result) String() string {
	if len(r.things) == 0 {
		return "No things found."
	}
	t := table.New()

	h := []interface{}{"Name", "ID", "Device", "Tags"}
	if r.variables {
		h = append(h, "Variables")
	}
	t.SetHeader(h...)

	for _, thing := range r.things {
		row := []interface{}{thing.Name, thing.ID, thing.DeviceID}
		row = append(row, strings.Join(thing.Tags, ","))
		if r.variables {
			row = append(row, strings.Join(thing.Variables, ", "))
		}
		t.AddRow(row...)
	}
	return t.Render()
}
