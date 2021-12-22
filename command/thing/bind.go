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

package thing

import (
	"github.com/arduino/arduino-cloud-cli/internal/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	iotclient "github.com/arduino/iot-client-go"
)

// BindParams contains the parameters needed to
// bind a thing to a device.
type BindParams struct {
	ID       string // ID of the thing to be bound
	DeviceID string // ID of the device to be bound
}

// Bind command is used to bind a thing to a device
// on Arduino IoT Cloud.
func Bind(params *BindParams) error {
	conf, err := config.RetrieveCredentials()
	if err != nil {
		return err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return err
	}

	thing := &iotclient.ThingUpdate{
		DeviceId: params.DeviceID,
	}

	err = iotClient.ThingUpdate(params.ID, thing, true)
	if err != nil {
		return err
	}

	return nil
}
