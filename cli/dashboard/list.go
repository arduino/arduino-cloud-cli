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
	"math"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/arduino-cloud-cli/command/dashboard"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	widgetsPerRow = 3
)

type listFlags struct {
	showWidgets bool
}

func initListCommand() *cobra.Command {
	flags := &listFlags{}
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List dashboards",
		Long:  "List dashboards on Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runListCommand(flags); err != nil {
				feedback.Errorf("Error during dashboard list: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}

	listCommand.Flags().BoolVarP(&flags.showWidgets, "show-widgets", "s", false, "Show names of dashboard widgets")
	return listCommand
}

func runListCommand(flags *listFlags) error {
	logrus.Info("Listing dashboards")

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	dash, err := dashboard.List(context.TODO(), cred)
	if err != nil {
		return err
	}

	feedback.PrintResult(listResult{dashboards: dash, showWidgets: flags.showWidgets})
	return nil
}

type listResult struct {
	dashboards  []dashboard.DashboardInfo
	showWidgets bool
}

func (r listResult) Data() interface{} {
	return r.dashboards
}

func (r listResult) String() string {
	if len(r.dashboards) == 0 {
		return "No dashboard found."
	}
	t := table.New()

	head := []interface{}{"Name", "ID", "UpdatedAt"}
	if r.showWidgets {
		head = append(head, "Widgets")
	}
	t.SetHeader(head...)

	for _, dash := range r.dashboards {
		row := []interface{}{dash.Name, dash.ID, dash.UpdatedAt}

		if r.showWidgets {
			// Limit number of widgets per row.
			if len(dash.Widgets) > widgetsPerRow {
				row = append(row, strings.Join(dash.Widgets[:widgetsPerRow], ", "))
				dash.Widgets = dash.Widgets[widgetsPerRow:]
			} else {
				row = append(row, strings.Join(dash.Widgets, ", "))
				dash.Widgets = nil
			}
		}
		t.AddRow(row...)

		// Print remaining widgets in new rows
		if r.showWidgets {
			for len(dash.Widgets) > 0 {
				row := []interface{}{"", "", ""}
				l := int(math.Min(float64(len(dash.Widgets)), widgetsPerRow))
				row = append(row, strings.Join(dash.Widgets[:l], ", "))
				dash.Widgets = dash.Widgets[l:]
				t.AddRow(row...)
			}
		}
	}
	return t.Render()
}
