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

package template

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	iotclient "github.com/arduino/iot-client-go/v3"
	"gopkg.in/yaml.v3"
)

// FromThing extracts a template of type map[string]interface{} from a thing.
func FromThing(thing *iotclient.ArduinoThing) map[string]interface{} {
	template := make(map[string]interface{})
	template["name"] = thing.Name
	template["timezone"] = thing.Timezone

	// Extract template from thing structure
	var props []map[string]interface{}
	for _, p := range thing.Properties {
		prop := make(map[string]interface{})
		prop["name"] = p.Name
		prop["permission"] = p.Permission
		prop["type"] = p.Type
		prop["update_parameter"] = p.UpdateParameter
		prop["update_strategy"] = p.UpdateStrategy
		prop["variable_name"] = p.VariableName
		props = append(props, prop)
	}
	template["variables"] = props

	if thing.Tags != nil {
		tags := []map[string]any{}
		for k, v := range thing.Tags {
			tag := make(map[string]any)
			tag[k] = v
			tags = append(tags, tag)
		}
		template["tags"] = tags
	}

	return template
}

func FromDashboard(dashboard *iotclient.ArduinoDashboardv3template) (*DashboardTemplate, error) {
	template := &DashboardTemplate{
		Name: dashboard.Name,
	}

	jsonDashboard, err := json.Marshal(dashboard)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal dashboard: %w", err)
	}

	err = json.Unmarshal(jsonDashboard, &template)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal dashboard: %w", err)
	}

	return template, nil
}

// ToFile takes a generic template and saves it into a file,
// in the specified format (yaml or json).
func ToFile(template map[string]interface{}, outfile string, format string) error {
	var file []byte
	var err error

	if format == "json" {
		file, err = json.MarshalIndent(template, "", "    ")
		if err != nil {
			return fmt.Errorf("%s: %w", "template marshal failure: ", err)
		}

	} else if format == "yaml" {
		file, err = yaml.Marshal(template)
		if err != nil {
			return fmt.Errorf("%s: %w", "template marshal failure: ", err)
		}

	} else {
		return errors.New("format is not valid: only 'json' and 'yaml' are supported")
	}

	err = os.WriteFile(outfile, file, os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("%s: %w", "cannot write outfile: ", err)
	}

	return nil
}
