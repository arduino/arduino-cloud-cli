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

package dashboard

import (
	"fmt"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/dashboard"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var createFlags struct {
	name     string
	template string
	override map[string]string
}

func initCreateCommand() *cobra.Command {
	createCommand := &cobra.Command{
		Use:   "create",
		Short: "Create a dashboard from a template",
		Long:  "Create a dashboard from a template for Arduino IoT Cloud",
		Run:   runCreateCommand,
	}
	createCommand.Flags().StringVarP(&createFlags.name, "name", "n", "", "Dashboard name")
	createCommand.Flags().StringVarP(&createFlags.template, "template", "t", "",
		"File containing a dashboard template, JSON and YAML format are supported",
	)
	createCommand.Flags().StringToStringVarP(&createFlags.override, "override", "o", nil,
		"Map stating the items to be overridden. Ex: 'thing-0=xxxxxxxx,thing-1=yyyyyyyy'")

	createCommand.MarkFlagRequired("template")
	return createCommand
}

func runCreateCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Creating dashboard from template %s", createFlags.template)

	params := &dashboard.CreateParams{
		Template: createFlags.template,
		Override: createFlags.override,
	}
	if createFlags.name != "" {
		params.Name = &createFlags.name
	}

	dashboard, err := dashboard.Create(params)
	if err != nil {
		feedback.Errorf("Error during dashboard create: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.PrintResult(createResult{dashboard})
}

type createResult struct {
	dashboard *dashboard.DashboardInfo
}

func (r createResult) Data() interface{} {
	return r.dashboard
}

func (r createResult) String() string {
	return fmt.Sprintf(
		"name: %s\nid: %s\nupdated_at: %s\nwidgets: %s",
		r.dashboard.Name,
		r.dashboard.ID,
		r.dashboard.UpdatedAt,
		strings.Join(r.dashboard.Widgets, ", "),
	)
}
