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

// CloneParams contains the parameters needed to clone a thing.
type CloneParams struct {
	Name    string // Name of the new thing
	CloneID string // ID of thing to be cloned
}

// Clone allows to create a new thing from an already existing one.
func Clone(ctx context.Context, params *CloneParams, cred *config.Credentials) (*ThingInfo, error) {
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return nil, err
	}

	newThing, err := iotClient.ThingClone(ctx, params.CloneID, params.Name)
	if err != nil {
		return nil, fmt.Errorf("cloning thing %s: %w", params.CloneID, err)
	}

	t, err := getThingInfo(newThing)
	if err != nil {
		return nil, fmt.Errorf("parsing thing %s from cloud: %w", newThing.Id, err)
	}
	return t, nil
}
