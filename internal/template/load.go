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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	iotclient "github.com/arduino/iot-client-go"
	"github.com/gofrs/uuid"
	"gopkg.in/yaml.v3"
)

// loadTemplate loads a template file and unmarshals it into whatever
// is pointed to by the template parameter. If template is nil or
// not a pointer, loadTemplate returns an error.
// file: path of a template file in json or yaml format.
func loadTemplate(file string, template interface{}) error {
	templateFile, err := os.Open(file)
	if err != nil {
		return err
	}
	defer templateFile.Close()

	templateBytes, err := ioutil.ReadAll(templateFile)
	if err != nil {
		return err
	}

	// Extract template trying all the supported formats: json and yaml
	if err = json.Unmarshal([]byte(templateBytes), template); err != nil {
		if err = yaml.Unmarshal([]byte(templateBytes), template); err != nil {
			return errors.New("reading template file: template format is not valid")
		}
	}

	return nil
}

// LoadThing loads a thing from a thing template file.
func LoadThing(file string) (*iotclient.ThingCreate, error) {
	var template map[string]interface{}
	err := loadTemplate(file, &template)
	if err != nil {
		return nil, err
	}

	// Adapt thing template to thing structure
	delete(template, "id")
	template["properties"] = template["variables"]
	delete(template, "variables")

	// Convert template into thing structure exploiting json marshalling/unmarshalling
	thing := &iotclient.ThingCreate{}

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

// LoadDashboard loads a dashboard from a dashboard template file.
// It applies the thing overrides specified by the override parameter.
// It requires a ThingFetcher to retrieve the actual variable ids.
func LoadDashboard(ctx context.Context, file string, override map[string]string, thinger ThingFetcher) (*iotclient.Dashboardv2, error) {
	template := dashboardTemplate{}
	err := loadTemplate(file, &template)
	if err != nil {
		return nil, err
	}

	// Adapt the template to the dashboard struct
	for i, widget := range template.Widgets {
		// Generate and set a uuid for each widget
		id, err := uuid.NewV4()
		if err != nil {
			return nil, fmt.Errorf("cannot create a uuid for new widget: %w", err)
		}
		widget.Id = id.String()
		filterWidgetOptions(widget.Options)
		// Even if the widget has no options, its field should exist
		if widget.Options == nil {
			widget.Options = make(map[string]interface{})
		}
		// Set the correct variable id, given the thing id and the variable name
		for j, variable := range widget.Variables {
			// Check if thing name should be overridden
			if id, ok := override[variable.ThingID]; ok {
				variable.ThingID = id
			}
			variable.VariableID, err = getVariableID(ctx, variable.ThingID, variable.VariableName, thinger)
			if err != nil {
				return nil, err
			}
			widget.Variables[j] = variable
		}
		template.Widgets[i] = widget
	}

	// Convert template into dashboard structure exploiting json marshalling/unmarshalling
	dashboard := &iotclient.Dashboardv2{}
	t, err := json.Marshal(template)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "extracting template", err)
	}
	err = json.Unmarshal(t, &dashboard)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "creating dashboard structure from template", err)
	}

	return dashboard, nil
}
