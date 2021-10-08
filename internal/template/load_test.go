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
	"testing"

	iotclient "github.com/arduino/iot-client-go"
	"github.com/google/go-cmp/cmp"
)

const (
	uuidv4Length = 36
)

var (
	dashboardTemplate = map[string]interface{}{
		"id":   "home-security-alarm-dashboard",
		"name": "Home Security Alarm",
		"widgets": []interface{}{
			map[string]interface{}{
				"type": "Messenger", "name": "message_update",
				"variables": []interface{}{map[string]interface{}{"thing_id": "home-security-alarm", "variable_id": "message_update"}},
			},
			map[string]interface{}{
				"type": "Switch", "name": "light_alarm",
				"variables": []interface{}{map[string]interface{}{"thing_id": "home-security-alarm", "variable_id": "light_alarm"}},
				"options":   map[string]interface{}{"showLabels": true},
			},
		},
	}

	dashboardDetailed = &iotclient.Dashboardv2{
		Name: "dashboard",
		Widgets: []iotclient.Widget{
			{Name: "Switch-name", Height: 1, HeightMobile: 2, Width: 3, WidthMobile: 4,
				X: 5, XMobile: 6, Y: 7, YMobile: 8, Options: map[string]interface{}{"showLabels": true},
				Type: "Switch",
			},
		},
	}

	dashboardNoOptions = &iotclient.Dashboardv2{
		Name: "dashboard-no-options",
		Widgets: []iotclient.Widget{
			{Name: "Switch-name", Height: 1, HeightMobile: 2, Width: 3, WidthMobile: 4,
				X: 5, XMobile: 6, Y: 7, YMobile: 8, Options: map[string]interface{}{},
				Type: "Switch",
			},
		},
	}

	dashboardWithVariable = &iotclient.Dashboardv2{
		Name: "dashboard-with-variable",
		Widgets: []iotclient.Widget{
			{Name: "Switch-name", Height: 1, HeightMobile: 2, Width: 3, WidthMobile: 4,
				X: 5, XMobile: 6, Y: 7, YMobile: 8, Options: map[string]interface{}{"showLabels": true},
				// in this test, the variable id is a concatenation of thing_id and variable_id
				// this depends on the mocked function getVariableID
				Type: "Switch", Variables: []string{"thing-variable"},
			},
		},
	}

	dashboardVariableOverride = &iotclient.Dashboardv2{
		Name: "dashboard-with-variable",
		Widgets: []iotclient.Widget{
			{Name: "Switch-name", Height: 1, HeightMobile: 2, Width: 3, WidthMobile: 4,
				X: 5, XMobile: 6, Y: 7, YMobile: 8, Options: map[string]interface{}{"showLabels": true},
				// in this test, the variable id is a concatenation of thing_id and variable_id
				// this depends on the mocked function getVariableID
				Type: "Switch", Variables: []string{"overridden-variable"},
			},
		},
	}

	dashboardTwoWidgets = &iotclient.Dashboardv2{
		Name: "dashboard-two-widgets",
		Widgets: []iotclient.Widget{
			{Name: "blink_speed", Height: 7, Width: 8,
				X: 7, Y: 5, Options: map[string]interface{}{"min": float64(0), "max": float64(5000)},
				Type: "Slider", Variables: []string{"remote-controlled-lights-blink_speed"},
			},
			{Name: "relay_2", Height: 5, Width: 5,
				X: 5, Y: 0, Options: map[string]interface{}{"showLabels": true},
				Type: "Switch", Variables: []string{"remote-controlled-lights-relay_2"},
			},
		},
	}
)

func TestLoadTemplate(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		override map[string]string
		want     map[string]interface{}
	}{

		{
			name: "yaml dashboard template",
			file: "testdata/home-security-dashboard.yaml",
			want: dashboardTemplate,
		},

		{
			name: "json dashboard template",
			file: "testdata/home-security-dashboard.json",
			want: dashboardTemplate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got map[string]interface{}
			err := loadTemplate(tt.file, &got)
			if err != nil {
				t.Errorf("%v", err)
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("Wrong template received, got=\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func TestLoadDashboard(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		override map[string]string
		want     *iotclient.Dashboardv2
	}{
		{
			name:     "dashboard detailed",
			file:     "testdata/dashboard-detailed.yaml",
			override: nil,
			want:     dashboardDetailed,
		},

		{
			name:     "dashboard with wrong options to be filtered out",
			file:     "testdata/dashboard-wrong-options.yaml",
			override: nil,
			want:     dashboardDetailed,
		},

		{
			name:     "dashboard without options, should have a not nil map",
			file:     "testdata/dashboard-no-options.yaml",
			override: nil,
			want:     dashboardNoOptions,
		},

		{
			name:     "dashboard with variable, mocked variable id is concatenation of thing_id and variable_id",
			file:     "testdata/dashboard-with-variable.yaml",
			override: nil,
			want:     dashboardWithVariable,
		},

		{
			name:     "dashboard with variable, thing is overridden",
			file:     "testdata/dashboard-with-variable.yaml",
			override: map[string]string{"thing": "overridden"},
			want:     dashboardVariableOverride,
		},

		{
			name:     "dashboard with two widgets",
			file:     "testdata/dashboard-two-widgets.yaml",
			override: nil,
			want:     dashboardTwoWidgets,
		},
	}

	vargetter.getVariableID = mockGetVariableID

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadDashboard(tt.file, tt.override)
			if err != nil {
				t.Errorf("%v", err)
			}

			for i := range got.Widgets {
				// check widget id generation
				id := got.Widgets[i].Id
				if len(id) != uuidv4Length {
					t.Errorf("Widget ID is wrong: = %s", id)
				}
				got.Widgets[i].Id = ""
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("Wrong template received, got=\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}

func mockGetVariableID(thingID string, variableName string) (string, error) {
	return thingID + "-" + variableName, nil
}
