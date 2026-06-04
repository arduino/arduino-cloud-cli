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
	"encoding/json"
	"strconv"

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	iotAPI "github.com/arduino/iot-client-go/v3"
)

func MultiplePublish(ctx context.Context, thingID string, values map[string]string, cred *config.Credentials) error {
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return err
	}

	properties := make([]iotAPI.PropertyDefinition, 0, len(values))
	for propertyName, value := range values {
		prop := iotAPI.PropertyDefinition{
			Property: propertyName,
			Value:    adaptType(value),
		}
		properties = append(properties, prop)
	}

	err = iotClient.MultiplePropertyPublish(ctx, thingID, properties)
	return err

}

func adaptType(s string) interface{} {
	// int
	if val, err := strconv.Atoi(s); err == nil {
		return val
	}

	// float
	if val, err := strconv.ParseFloat(s, 64); err == nil {
		return val
	}

	// bool
	if val, err := strconv.ParseBool(s); err == nil {
		return val
	}

	// JSON (oggetto o array)
	var js interface{}
	if err := json.Unmarshal([]byte(s), &js); err == nil {
		return js
	}

	return s
}
