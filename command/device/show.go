// This file is part of arduino-cloud-cli.
//
// Copyright (C) 2024 ARDUINO SA (http://www.arduino.cc/)
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

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	"github.com/sirupsen/logrus"
)

type netCredentials struct {
	FriendlyName string `json:"friendly_name"`
	Required     bool   `json:"required"`
	SecretName   string `json:"secret_name"`
	Sensitive    bool   `json:"sensitive"`
}

// List command is used to list
// the devices of Arduino IoT Cloud.
func Show(ctx context.Context, deviceId string, cred *config.Credentials) (*DeviceInfo, []netCredentials, error) {
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return nil, nil, err
	}

	logrus.Infof("Show device %s", deviceId)

	foundDevice, err := iotClient.DeviceShow(ctx, deviceId)
	if err != nil {
		return nil, nil, err
	}
	device, err := getDeviceInfo(foundDevice)
	if err != nil {
		return nil, nil, err
	}

	net := []netCredentials{}
	if device.ConnectionType != nil {
		netCredentialsArray, err := iotClient.DeviceNetworkCredentials(ctx, deviceId, *foundDevice.ConnectionType)
		if err != nil {
			return nil, net, err
		}
		for _, netCred := range netCredentialsArray {
			var netCredToShow netCredentials
			netCredToShow.FriendlyName = netCred.FriendlyName
			netCredToShow.Required = netCred.Required
			netCredToShow.SecretName = netCred.SecretName
			netCredToShow.Sensitive = netCred.Sensitive
			net = append(net, netCredToShow)
		}
	}

	return device, nil, nil
}
