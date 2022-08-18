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
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/arduino-cloud-cli/command/device"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initListFQBNCommand() *cobra.Command {
	listFQBNCommand := &cobra.Command{
		Use:   "list-fqbn",
		Short: "List supported FQBN",
		Long:  "List all the FQBN supported by Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runListFQBNCommand(); err != nil {
				feedback.Errorf("Error during device list-fqbn: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	return listFQBNCommand
}

func runListFQBNCommand() error {
	logrus.Info("Listing supported FQBN")

	fqbn, err := device.ListFQBN(context.TODO())
	if err != nil {
		return err
	}

	feedback.PrintResult(listFQBNResult{fqbn})
	return nil
}

type listFQBNResult struct {
	fqbn []device.FQBNInfo
}

func (r listFQBNResult) Data() interface{} {
	return r.fqbn
}

func (r listFQBNResult) String() string {
	if len(r.fqbn) == 0 {
		return "No FQBN."
	}
	t := table.New()
	t.SetHeader("Name", "FQBN")
	for _, f := range r.fqbn {
		t.AddRow(
			f.Name,
			f.Value,
		)
	}
	return t.Render()
}
