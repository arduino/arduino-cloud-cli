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
	"context"
	"encoding/json"
	"fmt"

	iotclient "github.com/arduino/iot-client-go/v3"
)

type DashboardTemplate struct {
	ID      string           `json:"id" yaml:"id"`
	Name    string           `json:"name,omitempty" yaml:"name,omitempty"`
	Pages   []Page           `json:"pages,omitempty" yaml:"pages,omitempty"`
	Widgets []WidgetTemplate `json:"widgets,omitempty" yaml:"widgets,omitempty"`
}

type Page struct {
	Id       string  `json:"id" yaml:"id"`
	Name     string  `json:"name" yaml:"name"`
	Position int     `json:"position" yaml:"position"`
	Icon     *string `json:"icon,omitempty" yaml:"icon,omitempty"`
}

type WidgetTemplate struct {
	Id           string                 `json:"id,omitempty" yaml:"id,omitempty"`
	Type         string                 `json:"type" yaml:"type"`
	Name         string                 `json:"name" yaml:"name"`
	Width        int                    `json:"width" yaml:"width"`
	Height       int                    `json:"height" yaml:"height"`
	X            int                    `json:"x" yaml:"x"`
	Y            int                    `json:"y" yaml:"y"`
	WidthMobile  *int                   `json:"width_mobile" yaml:"width_mobile"`
	HeightMobile *int                   `json:"height_mobile" yaml:"height_mobile"`
	XMobile      *int                   `json:"x_mobile" yaml:"x_mobile"`
	YMobile      *int                   `json:"y_mobile" yaml:"y_mobile"`
	Variables    []VariableTemplate     `json:"variables" yaml:"variables"`
	PageID       string                 `json:"page_id" yaml:"page_id"`
	Options      map[string]interface{} `json:"options" yaml:"options"`
}

type VariableTemplate struct {
	ThingID    string `json:"thing_id" yaml:"thing_id"`
	VariableID string `json:"variable_id" yaml:"variable_id"`
	Name       string `json:"name" yaml:"name"`
}

// MarshalJSON satisfies the Marshaler interface from json package.
// With this, when a variableTemplate is marshaled, it only marshals
// its VariableID. In this way, a widgetTemplate can be
// marshaled and then unmarshaled into a iot.Widget struct.
// In the same way, a dashboardTemplate can now be converted
// into a iot.DashboardV2 leveraging the JSON marshal/unmarshal.
func (v *VariableTemplate) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.VariableID)
}

// ThingFetcher wraps the method to fetch a thing given its id.
type ThingFetcher interface {
	ThingShow(ctx context.Context, id string) (*iotclient.ArduinoThing, error)
	PropertyShow(ctx context.Context, thingId, variableId string) (*iotclient.ArduinoProperty, error)
}

// getVariableID returns the id of a variable, given its name and its thing id.
// If the variable is not found, an error is returned.
func getVariableID(ctx context.Context, thingID string, variableName string, fetcher ThingFetcher) (string, error) {
	thing, err := fetcher.ThingShow(ctx, thingID)
	if err != nil {
		return "", fmt.Errorf("getting variables of thing %s: %w", thingID, err)
	}

	for _, v := range thing.Properties {
		if v.VariableName != nil && *v.VariableName == variableName {
			return v.Id, nil
		}
	}

	return "", fmt.Errorf("thing with id %s doesn't have variable with name %s : %w", thingID, variableName, err)
}
