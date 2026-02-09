// This file is part of arduino-cloud-cli.
//
// Copyright (C) ARDUINO SRL (http://www.arduino.cc/)
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
	iotclient "github.com/arduino/iot-client-go/v3"
)

// DashboardInfo contains the most interesting
// information, in string format, of an Arduino IoT Cloud dashboard.
type DashboardInfo struct {
	Name      string   `json:"name"`
	ID        string   `json:"id"`
	UpdatedAt string   `json:"updated_at"`
	Widgets   []string `json:"widgets"`
}

func getDashboardInfo(dashboard *iotclient.ArduinoDashboardv2) *DashboardInfo {
	var widgets []string
	for _, w := range dashboard.Widgets {
		if w.Name != nil {
			widgets = append(widgets, *w.Name)
		}
	}
	info := &DashboardInfo{
		Name:      dashboard.Name,
		ID:        dashboard.Id,
		UpdatedAt: dashboard.UpdatedAt.String(),
		Widgets:   widgets,
	}
	return info
}
