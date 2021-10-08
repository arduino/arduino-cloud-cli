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

package template

import (
	"errors"

	"github.com/arduino/arduino-cloud-cli/internal/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
)

// HACK: this global variable is used to mock getVariableID function during tests.
// This method is temporarily.
var vargetter = varGetter{
	getVariableID: getVariableID,
}

type varGetter struct {
	getVariableID func(thingID string, variableName string) (string, error)
}

// inefficient: creates a new iot client for each variable name
// solutions: pass the client from the extern. this solves also test problems
// instantiate the client as a state of vargetter
func getVariableID(thingID string, variableName string) (string, error) {
	conf, err := config.Retrieve()
	if err != nil {
		return "", err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return "", err
	}

	thing, err := iotClient.ThingShow(thingID)
	if err != nil {
		return "", err
	}

	for _, v := range thing.Properties {
		if v.Name == variableName {
			return v.Id, nil
		}
	}

	return "", errors.New("not found")
}
