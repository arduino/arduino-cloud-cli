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
	"fmt"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/thing"
	"github.com/arduino/arduino-cloud-cli/internal/config"
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
		Run:   runCreateCommand,
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

func runCreateCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Creating thing from template %s", createFlags.template)

	cred, err := config.RetrieveCredentials()
	if err != nil {
		feedback.Errorf("Error during thing create: retrieving credentials: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	params := &thing.CreateParams{
		Template: createFlags.template,
	}
	if createFlags.name != "" {
		params.Name = &createFlags.name
	}

	thing, err := thing.Create(params, cred)
	if err != nil {
		feedback.Errorf("Error during thing create: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.PrintResult(createResult{thing})
}

type createResult struct {
	thing *thing.ThingInfo
}

func (r createResult) Data() interface{} {
	return r.thing
}

func (r createResult) String() string {
	return fmt.Sprintf(
		"name: %s\nid: %s\ndevice_id: %s\nvariables: %s",
		r.thing.Name,
		r.thing.ID,
		r.thing.DeviceID,
		strings.Join(r.thing.Variables, ", "),
	)
}
