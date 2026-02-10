// This file is part of arduino-cloud-cli.
//
// Copyright (C) ARDUINO SRL (http://www.arduino.cc)
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
	"github.com/arduino/arduino-cloud-cli/command/tag"
	iotclient "github.com/arduino/iot-client-go/v3"
)

// DeviceInfo contains the most interesting
// parameters of an Arduino IoT Cloud device.
type DeviceInfo struct {
	Name           string   `json:"name"`
	ID             string   `json:"id"`
	Board          string   `json:"board"`
	Serial         string   `json:"serial_number"`
	FQBN           string   `json:"fqbn"`
	Tags           []string `json:"tags,omitempty"`
	Status         *string  `json:"status,omitempty"`
	Type           string   `json:"type,omitempty"`
	ConnectionType *string  `json:"connection_type,omitempty"`
	ThingID        *string  `json:"thing_id,omitempty"`
}

func getDeviceInfo(device *iotclient.ArduinoDevicev2) (*DeviceInfo, error) {
	// Retrieve device tags
	tags, err := tag.TagsInfo(device.Tags)
	if err != nil {
		return nil, err
	}

	dev := &DeviceInfo{
		Name:           device.Name,
		ID:             device.Id,
		Board:          device.Type,
		Serial:         device.Serial,
		FQBN:           dereferenceString(device.Fqbn),
		Tags:           tags,
		Status:         device.DeviceStatus,
		Type:           device.Type,
		ConnectionType: device.ConnectionType,
	}
	if device.Thing != nil {
		dev.ThingID = &device.Thing.Id
	}
	return dev, nil
}
