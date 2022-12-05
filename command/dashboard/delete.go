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

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
)

// DeleteParams contains the parameters needed to
// delete a dashboard from Arduino IoT Cloud.
type DeleteParams struct {
	ID string
}

// Delete command is used to delete a dashboard
// from Arduino IoT Cloud.
func Delete(ctx context.Context, params *DeleteParams, cred *config.Credentials) error {
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return err
	}

	return iotClient.DashboardDelete(ctx, params.ID)
}
