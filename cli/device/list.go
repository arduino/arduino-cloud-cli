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

package device

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/arduino-cloud-cli/command/device"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type listFlags struct {
	tags map[string]string
}

func initListCommand() *cobra.Command {
	flags := &listFlags{}
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List devices",
		Long:  "List devices on Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runListCommand(flags); err != nil {
				feedback.Errorf("Error during device list: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	listCommand.Flags().StringToStringVar(
		&flags.tags,
		"tags",
		nil,
		"Comma-separated list of tags with format <key>=<value>.\n"+
			"List only devices that match the provided tags.",
	)
	return listCommand
}

func runListCommand(flags *listFlags) error {
	logrus.Info("Listing devices")

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	params := &device.ListParams{Tags: flags.tags}
	devs, err := device.List(context.TODO(), params, cred)
	if err != nil {
		return err
	}

	feedback.PrintResult(listResult{devs})
	return nil
}

type listResult struct {
	devices []device.DeviceInfo
}

func (r listResult) Data() interface{} {
	return r.devices
}

func (r listResult) String() string {
	if len(r.devices) == 0 {
		return "No devices found."
	}
	t := table.New()
	t.SetHeader("Name", "ID", "Board", "FQBN", "SerialNumber", "Tags")
	for _, device := range r.devices {
		t.AddRow(
			device.Name,
			device.ID,
			device.Board,
			device.FQBN,
			device.Serial,
			strings.Join(device.Tags, ","),
		)
	}
	return t.Render()
}
