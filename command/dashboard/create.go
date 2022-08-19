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
	"errors"

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	"github.com/arduino/arduino-cloud-cli/internal/template"
)

// CreateParams contains the parameters needed to create a new dashboard.
type CreateParams struct {
	Name     *string           // Name of the new dashboard
	Override map[string]string // Template parameters to be overridden
	Template string            // Path of the template file
}

// Create allows to create a new dashboard.
func Create(params *CreateParams, cred *config.Credentials) (*DashboardInfo, error) {
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return nil, err
	}

	dashboard, err := template.LoadDashboard(params.Template, params.Override, iotClient)
	if err != nil {
		return nil, err
	}

	// Name passed as parameter has priority over name from template
	if params.Name != nil {
		dashboard.Name = *params.Name
	}
	// If name is not specified in the template, it should be passed as parameter
	if dashboard.Name == "" {
		return nil, errors.New("dashboard name not specified")
	}

	newDashboard, err := iotClient.DashboardCreate(dashboard)
	if err != nil {
		return nil, err
	}

	return getDashboardInfo(newDashboard), nil
}
