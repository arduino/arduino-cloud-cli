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
	"math"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/arduino-cloud-cli/command/dashboard"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	widgetsPerRow = 3
)

var listFlags struct {
	showSharing bool
}

func initListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List dashboards",
		Long:  "List dashboards on Arduino IoT Cloud",
		Run:   runListCommand,
	}

	listCommand.Flags().BoolVarP(&listFlags.showSharing, "show-sharing", "s", false, "Show dashboard sharing information")
	return listCommand
}

func runListCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Listing dashboards")

	dash, err := dashboard.List()
	if err != nil {
		feedback.Errorf("Error during dashboard list: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.PrintResult(listResult{dash})
}

type listResult struct {
	dashboards []dashboard.DashboardInfo
}

func (r listResult) Data() interface{} {
	return r.dashboards
}

func (r listResult) String() string {
	if len(r.dashboards) == 0 {
		return "No dashboard found."
	}
	t := table.New()

	head := []interface{}{"Name", "ID", "Widgets", "UpdatedAt"}
	t.SetHeader(head...)

	for _, dash := range r.dashboards {
		row := []interface{}{dash.Name, dash.ID}

		// Limit number of widgets per row.
		if len(dash.Widgets) > widgetsPerRow {
			row = append(row, strings.Join(dash.Widgets[:widgetsPerRow], ", "))
			dash.Widgets = dash.Widgets[widgetsPerRow:]
		} else {
			row = append(row, strings.Join(dash.Widgets, ", "))
			dash.Widgets = nil
		}
		row = append(row, dash.UpdatedAt)
		t.AddRow(row...)

		// Print remaining widgets in new rows
		for len(dash.Widgets) > 0 {
			row := []interface{}{"", ""}
			l := int(math.Min(float64(len(dash.Widgets)), widgetsPerRow))
			row = append(row, strings.Join(dash.Widgets[:l], ", "))
			dash.Widgets = dash.Widgets[l:]
			row = append(row, "")
			t.AddRow(row...)
		}
	}
	return t.Render()
}
