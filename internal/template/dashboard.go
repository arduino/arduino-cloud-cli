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

import "encoding/json"

type dashboardHelp struct {
	Name    string       `json:"name,omitempty" yaml:"name,omitempty"`
	Widgets []widgetHelp `json:"widgets,omitempty" yaml:"widgets,omitempty"`
}

type widgetHelp struct {
	Height       int64                  `json:"height" yaml:"height"`
	HeightMobile int64                  `json:"height_mobile,omitempty" yaml:"height_mobile,omitempty"`
	Id           string                 `json:"id" yaml:"id"`
	Name         string                 `json:"name,omitempty" yaml:"name,omitempty"`
	Options      map[string]interface{} `json:"options" yaml:"options"`
	WidgetType   string                 `json:"type" yaml:"type"`
	Variables    []variableHelp         `json:"variables,omitempty" yaml:"variables,omitempty"`
	Width        int64                  `json:"width" yaml:"width"`
	WidthMobile  int64                  `json:"width_mobile,omitempty" yaml:"width_mobile,omitempty"`
	X            int64                  `json:"x" yaml:"x"`
	XMobile      int64                  `json:"x_mobile,omitempty" yaml:"x_mobile,omitempty"`
	Y            int64                  `json:"y" yaml:"y"`
	YMobile      int64                  `json:"y_mobile,omitempty" yaml:"y_mobile,omitempty"`
}

type variableHelp struct {
	ThingID      string `json:"thing_id" yaml:"thing_id"`
	VariableName string `json:"variable_id" yaml:"variable_id"`
	VariableID   string
}

func (v *variableHelp) MarshalJSON() ([]byte, error) {
	// Jsonize as a list of strings (variable uuids)
	// in order to uniform to the other dashboard declaration (of iotclient)
	return json.Marshal(v.VariableID)
}
