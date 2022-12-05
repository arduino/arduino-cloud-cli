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

package device

import (
	"context"
	"errors"

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
)

// DeleteParams contains the parameters needed to
// delete a device from Arduino IoT Cloud.
// ID and Tags parameters are mutually exclusive
// and one among them is required:  An error is returned
// if they are both nil or if they are both not nil.
type DeleteParams struct {
	ID   *string
	Tags map[string]string
}

// Delete command is used to delete a device
// from Arduino IoT Cloud.
func Delete(ctx context.Context, params *DeleteParams, cred *config.Credentials) error {
	if params.ID == nil && params.Tags == nil {
		return errors.New("provide either ID or Tags")
	} else if params.ID != nil && params.Tags != nil {
		return errors.New("cannot use both ID and Tags. only one of them should be not nil")
	}

	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return err
	}

	deviceIDs := []string{}
	if params.ID != nil {
		deviceIDs = append(deviceIDs, *params.ID)
	}
	if params.Tags != nil {
		dev, err := iotClient.DeviceList(ctx, params.Tags)
		if err != nil {
			return err
		}
		for _, d := range dev {
			deviceIDs = append(deviceIDs, d.Id)
		}
	}

	for _, id := range deviceIDs {
		err = iotClient.DeviceDelete(ctx, id)
		if err != nil {
			return err
		}
	}

	return nil
}
