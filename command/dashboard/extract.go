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

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	"github.com/arduino/arduino-cloud-cli/internal/template"
)

// ExtractParams contains the parameters needed to
// extract a template dashboard from Arduino IoT Cloud.
type ExtractParams struct {
	ID string
}

// Extract command is used to extract a dashboard template
// from a dashboard on Arduino IoT Cloud.
func Extract(ctx context.Context, params *ExtractParams, cred *config.Credentials) (map[string]interface{}, error) {
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return nil, err
	}

	dashboard, err := iotClient.DashboardShow(ctx, params.ID)
	if err != nil {
		err = fmt.Errorf("%s: %w", "cannot extract dashboard: ", err)
		return nil, err
	}

	return template.FromDashboard(dashboard), nil
}
