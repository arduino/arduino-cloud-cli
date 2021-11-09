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
	"github.com/arduino/arduino-cloud-cli/command/tag"
	iotclient "github.com/arduino/iot-client-go"
)

// ThingInfo contains the main parameters of
// an Arduino IoT Cloud thing.
type ThingInfo struct {
	Name      string   `json:"name"`
	ID        string   `json:"id"`
	DeviceID  string   `json:"device-id"`
	Variables []string `json:"variables"`
	Tags      []string `json:"tags,omitempty"`
}

func getThingInfo(thing *iotclient.ArduinoThing) (*ThingInfo, error) {
	// Retrieve thing variables
	var vars []string
	for _, p := range thing.Properties {
		vars = append(vars, p.Name)
	}
	// Retrieve thing tags
	tags, err := tag.Tags(thing.Tags).Info()
	if err != nil {
		return nil, err
	}

	info := &ThingInfo{
		Name:      thing.Name,
		ID:        thing.Id,
		DeviceID:  thing.DeviceId,
		Variables: vars,
		Tags:      tags,
	}
	return info, nil
}
