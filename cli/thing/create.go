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
	"github.com/arduino/arduino-cloud-cli/command/thing"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type createFlags struct {
	name     string
	template string
}

func initCreateCommand() *cobra.Command {
	flags := &createFlags{}
	createCommand := &cobra.Command{
		Use:   "create",
		Short: "Create a thing from a template",
		Long:  "Create a thing from a template for Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runCreateCommand(flags); err != nil {
				feedback.Errorf("Error during thing create: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	createCommand.Flags().StringVarP(&flags.name, "name", "n", "", "Thing name")
	createCommand.Flags().StringVarP(
		&flags.template,
		"template",
		"t",
		"",
		"File containing a thing template, JSON and YAML format are supported",
	)
	createCommand.MarkFlagRequired("template")
	return createCommand
}

func runCreateCommand(flags *createFlags) error {
	logrus.Infof("Creating thing from template %s", flags.template)

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	params := &thing.CreateParams{
		Template: flags.template,
	}
	if flags.name != "" {
		params.Name = &flags.name
	}

	thing, err := thing.Create(context.TODO(), params, cred)
	if err != nil {
		return err
	}

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
	return fmt.Sprintf(
		"name: %s\nid: %s\ndevice_id: %s\nvariables: %s",
		r.thing.Name,
		r.thing.ID,
		r.thing.DeviceID,
		strings.Join(r.thing.Variables, ", "),
	)
}
