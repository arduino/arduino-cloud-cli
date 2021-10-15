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

// FromThing extracts a template of type map[string]interface{} from a thing.
func FromThing(thing *iotclient.ArduinoThing) map[string]interface{} {
	template := make(map[string]interface{})
	template["name"] = thing.Name

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

	return template
}

// FromDashboard extracts a template of type map[string]interface{} from a dashboard.
func FromDashboard(dashboard *iotclient.ArduinoDashboardv2) map[string]interface{} {
	template := make(map[string]interface{})
	template["name"] = dashboard.Name

	// Extract template from dashboard structure
	var widgets []map[string]interface{}
	for _, w := range dashboard.Widgets {
		widget := make(map[string]interface{})
		widget["type"] = w.Type
		widget["name"] = w.Name
		widget["width"] = w.Width
		widget["height"] = w.Height
		widget["x"] = w.X
		widget["y"] = w.Y

		if w.WidthMobile != 0 && w.HeightMobile != 0 {
			widget["width_mobile"] = w.WidthMobile
			widget["height_mobile"] = w.HeightMobile
			widget["x_mobile"] = w.XMobile
			widget["y_mobile"] = w.YMobile
		}

		var vars []map[string]interface{}
		for _, v := range w.Variables {
			variable := make(map[string]interface{})
			variable["thing_id"] = v.ThingName
			variable["variable_id"] = v.VariableName
			vars = append(vars, variable)
		}
		if len(vars) > 0 {
			widget["variables"] = vars
		}

		filterWidgetOptions(w.Options)
		if len(w.Options) > 0 {
			widget["options"] = w.Options
		}
		widgets = append(widgets, widget)
	}
	if len(widgets) > 0 {
		template["widgets"] = widgets
	}
	return template
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

	err = ioutil.WriteFile(outfile, file, os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("%s: %w", "cannot write outfile: ", err)
	}

	return nil
}
