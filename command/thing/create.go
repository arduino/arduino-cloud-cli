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
	"errors"
	"fmt"

	"github.com/arduino/arduino-cloud-cli/internal/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	"github.com/arduino/arduino-cloud-cli/internal/template"
)

// CreateParams contains the parameters needed to create a new thing.
type CreateParams struct {
	Name     *string // Name of the new thing
	Template string  // Path of the template file
}

// Create allows to create a new thing.
func Create(params *CreateParams) (*ThingInfo, error) {
	conf, err := config.RetrieveCredentials()
	if err != nil {
		return nil, err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return nil, err
	}

	thing, err := template.LoadThing(params.Template)
	if err != nil {
		return nil, err
	}

	// Name passed as parameter has priority over name from template
	if params.Name != nil {
		thing.Name = *params.Name
	}
	// If name is not specified in the template, it should be passed as parameter
	if thing.Name == "" {
		return nil, errors.New("thing name not specified")
	}

	force := true
	newThing, err := iotClient.ThingCreate(thing, force)
	if err != nil {
		return nil, err
	}

	t, err := getThingInfo(newThing)
	if err != nil {
		return nil, fmt.Errorf("parsing the new thing %s from cloud: %w", newThing.Id, err)
	}
	return t, nil
}
