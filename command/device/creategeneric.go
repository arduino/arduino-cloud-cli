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
	"fmt"

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
)

const (
	genericDType = "login_and_secretkey_wifi"
)

// CreateGenericParams contains the parameters needed
// to create a new generic device.
type CreateGenericParams struct {
	Name string // Device name
	FQBN string // Board FQBN
}

// DeviceGenericInfo contains the most interesting
// parameters of a generic Arduino IoT Cloud device.
type DeviceGenericInfo struct {
	DeviceInfo
	Password string `json:"secret_key"`
}

// CreateGeneric command is used to add a new generic device to Arduino IoT Cloud.
func CreateGeneric(ctx context.Context, params *CreateGenericParams, cred *config.Credentials) (*DeviceGenericInfo, error) {
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return nil, err
	}

	dev, err := iotClient.DeviceCreate(context.Background(), params.FQBN, params.Name, "", genericDType)
	if err != nil {
		return nil, err
	}

	pass, err := iotClient.DevicePassSet(ctx, dev.Id)
	if err != nil {
		// Don't use the passed context for the cleanup because it could be cancelled.
		if errDel := iotClient.DeviceDelete(context.Background(), dev.Id); errDel != nil {
			return nil, fmt.Errorf(
				"device was successfully created on IoT-API but " +
					"now we can't set its secret key nor delete it - please check " +
					"it on the web application.\n\nFetch error: " + err.Error() +
					"\nDeletion error: " + errDel.Error(),
			)
		}
		return nil, fmt.Errorf("cannot create generic device: %w", err)
	}

	devInfo := &DeviceGenericInfo{
		DeviceInfo: DeviceInfo{
			Name:   dev.Name,
			ID:     dev.Id,
			Board:  dev.Type,
			Serial: dev.Serial,
			FQBN:   dev.Fqbn,
		},
		Password: pass.SuggestedPassword,
	}
	return devInfo, nil
}
