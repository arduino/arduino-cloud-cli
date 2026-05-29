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

package thing

import (
	"context"

	v3 "github.com/arduino/iot-client-go/v3"

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
)

// Show command is used to show
// a thing on Arduino IoT Cloud.
func Show(ctx context.Context, cred *config.Credentials, thingID string) (*v3.ArduinoThing, error) {
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return nil, err
	}

	thing, err := iotClient.ThingShow(ctx, thingID)
	if err != nil {
		return nil, err
	}

	return thing, nil
}
