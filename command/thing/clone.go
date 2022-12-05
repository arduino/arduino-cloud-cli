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
	iotclient "github.com/arduino/iot-client-go"
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

	thing, err := retrieve(ctx, iotClient, params.CloneID)
	if err != nil {
		return nil, err
	}

	thing.Name = params.Name
	force := true
	newThing, err := iotClient.ThingCreate(ctx, thing, force)
	if err != nil {
		return nil, err
	}

	t, err := getThingInfo(newThing)
	if err != nil {
		return nil, fmt.Errorf("parsing thing %s from cloud: %w", newThing.Id, err)
	}
	return t, nil
}

type thingFetcher interface {
	ThingShow(ctx context.Context, id string) (*iotclient.ArduinoThing, error)
}

func retrieve(ctx context.Context, fetcher thingFetcher, thingID string) (*iotclient.ThingCreate, error) {
	clone, err := fetcher.ThingShow(ctx, thingID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "retrieving the thing to be cloned", err)
	}

	thing := &iotclient.ThingCreate{}

	// Copy variables
	for _, p := range clone.Properties {
		thing.Properties = append(thing.Properties, iotclient.Property{
			Name:            p.Name,
			Permission:      p.Permission,
			UpdateParameter: p.UpdateParameter,
			UpdateStrategy:  p.UpdateStrategy,
			Type:            p.Type,
			VariableName:    p.VariableName,
		})
	}

	return thing, nil
}
