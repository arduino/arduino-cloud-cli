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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	iotclient "github.com/arduino/iot-client-go"
	"gopkg.in/yaml.v3"
)

// loadTemplate loads a template file and puts it into a generic template
// of type map[string]interface{}.
// The input template should be in json or yaml format.
func loadTemplate(file string) (map[string]interface{}, error) {
	templateFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer templateFile.Close()

	templateBytes, err := ioutil.ReadAll(templateFile)
	if err != nil {
		return nil, err
	}

	template := make(map[string]interface{})

	// Extract template trying all the supported formats: json and yaml
	if err = json.Unmarshal([]byte(templateBytes), &template); err != nil {
		if err = yaml.Unmarshal([]byte(templateBytes), &template); err != nil {
			return nil, errors.New("reading template file: template format is not valid")
		}
	}

	return template, nil
}

// LoadThing loads a thing from a thing template file.
func LoadThing(file string) (*iotclient.Thing, error) {
	template, err := loadTemplate(file)
	if err != nil {
		return nil, err
	}

	// Adapt thing template to thing structure
	delete(template, "id")
	template["properties"] = template["variables"]
	delete(template, "variables")

	// Convert template into thing structure exploiting json marshalling/unmarshalling
	thing := &iotclient.Thing{}

	t, err := json.Marshal(template)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "extracting template", err)
	}

	err = json.Unmarshal(t, &thing)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "creating thing structure from template", err)
	}

	return thing, nil
}
