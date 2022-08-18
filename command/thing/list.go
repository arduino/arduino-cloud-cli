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
	"context"
	"fmt"

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
)

// ListParams contains the optional parameters needed
// to filter the things to be listed.
type ListParams struct {
	IDs       []string          // If IDs is not nil, only things belonging to that list are returned
	DeviceID  *string           // If DeviceID is provided, only the thing associated to that device is listed.
	Variables bool              // If Variables is true, variable names are retrieved.
	Tags      map[string]string // If tags are provided, only things that have all these tags are listed.
}

// List command is used to list
// the things of Arduino IoT Cloud.
func List(ctx context.Context, params *ListParams, cred *config.Credentials) ([]ThingInfo, error) {
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return nil, err
	}

	foundThings, err := iotClient.ThingList(ctx, params.IDs, params.DeviceID, params.Variables, params.Tags)
	if err != nil {
		return nil, err
	}

	var things []ThingInfo
	for _, foundThing := range foundThings {
		info, err := getThingInfo(&foundThing)
		if err != nil {
			return nil, fmt.Errorf("parsing thing %s from cloud: %w", foundThing.Id, err)
		}
		things = append(things, *info)
	}

	return things, nil
}
