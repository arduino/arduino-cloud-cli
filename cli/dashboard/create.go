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
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/dashboard"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type createFlags struct {
	name     string
	template string
	override map[string]string
}

func initCreateCommand() *cobra.Command {
	flags := &createFlags{}
	createCommand := &cobra.Command{
		Use:   "create",
		Short: "Create a dashboard from a template",
		Long:  "Create a dashboard from a template for Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runCreateCommand(flags); err != nil {
				feedback.Errorf("Error during dashboard create: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	createCommand.Flags().StringVarP(&flags.name, "name", "n", "", "Dashboard name")
	createCommand.Flags().StringVarP(&flags.template, "template", "t", "",
		"File containing a dashboard template, JSON and YAML format are supported",
	)
	createCommand.Flags().StringToStringVarP(&flags.override, "override", "o", nil,
		"Map stating the items to be overridden. Ex: 'thing-0=xxxxxxxx,thing-1=yyyyyyyy'")

	createCommand.MarkFlagRequired("template")
	return createCommand
}

func runCreateCommand(flags *createFlags) error {
	logrus.Infof("Creating dashboard from template %s", flags.template)

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	params := &dashboard.CreateParams{
		Template: flags.template,
		Override: flags.override,
	}
	if flags.name != "" {
		params.Name = &flags.name
	}

	dashboard, err := dashboard.Create(context.TODO(), params, cred)
	if err != nil {
		return err
	}

	feedback.PrintResult(createResult{dashboard})
	return nil
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
