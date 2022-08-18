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
	"testing"

	iotclient "github.com/arduino/iot-client-go"
	"github.com/gofrs/uuid"
	"github.com/google/go-cmp/cmp"
)

const (
	// Real IDs will be UUIDs v4 like this: 9231a50b-8680-4489-a465-2b769fc310cb
	// Here we use these text strings to improve test errors readability
	switchyID    = "switchy-id"
	relayID      = "relay-id"
	blinkSpeedID = "blink_speed-id"

	thingOverriddenID   = "thing-overridden-id"
	switchyOverriddenID = "switchy-overridden-id"
)

var (
	dashboardTemplateTest = map[string]interface{}{
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
				X: 5, XMobile: 6, Y: 7, YMobile: 8, Options: map[string]interface{}{"showLabels": true}, Type: "Switch",
				Variables: []string{switchyID},
			},
		},
	}

	dashboardVariableOverride = &iotclient.Dashboardv2{
		Name: "dashboard-with-variable",
		Widgets: []iotclient.Widget{
			{Name: "Switch-name", Height: 1, HeightMobile: 2, Width: 3, WidthMobile: 4,
				X: 5, XMobile: 6, Y: 7, YMobile: 8, Options: map[string]interface{}{"showLabels": true}, Type: "Switch",
				Variables: []string{switchyOverriddenID},
			},
		},
	}

	dashboardTwoWidgets = &iotclient.Dashboardv2{
		Name: "dashboard-two-widgets",
		Widgets: []iotclient.Widget{
			{Name: "blink_speed", Height: 7, Width: 8,
				X: 7, Y: 5, Options: map[string]interface{}{"min": float64(0), "max": float64(5000)}, Type: "Slider",
				Variables: []string{blinkSpeedID},
			},
			{Name: "relay_2", Height: 5, Width: 5,
				X: 5, Y: 0, Options: map[string]interface{}{"showLabels": true}, Type: "Switch",
				Variables: []string{relayID},
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
			want: dashboardTemplateTest,
		},

		{
			name: "json dashboard template",
			file: "testdata/home-security-dashboard.json",
			want: dashboardTemplateTest,
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
				t.Errorf("Wrong template received, diff:\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}

type thingShowTest struct{}

func (t *thingShowTest) ThingShow(ctx context.Context, thingID string) (*iotclient.ArduinoThing, error) {
	if thingID == thingOverriddenID {
		return &iotclient.ArduinoThing{
			Properties: []iotclient.ArduinoProperty{
				{Id: switchyOverriddenID, Name: "switchy"},
			},
		}, nil
	}
	return &iotclient.ArduinoThing{
		Properties: []iotclient.ArduinoProperty{
			{Id: switchyID, Name: "switchy"},
			{Id: relayID, Name: "relay_2"},
			{Id: blinkSpeedID, Name: "blink_speed"},
		},
	}, nil
}

func TestLoadDashboard(t *testing.T) {
	mockThingShow := &thingShowTest{}
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
			override: map[string]string{"thing": thingOverriddenID},
			want:     dashboardVariableOverride,
		},

		{
			name:     "dashboard with two widgets",
			file:     "testdata/dashboard-two-widgets.yaml",
			override: nil,
			want:     dashboardTwoWidgets,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadDashboard(context.TODO(), tt.file, tt.override, mockThingShow)
			if err != nil {
				t.Errorf("%v", err)
			}

			for i := range got.Widgets {
				// check widget id generation
				id := got.Widgets[i].Id
				_, err := uuid.FromString(id)
				if err != nil {
					t.Errorf("Widget ID is not a valid UUID: %s", id)
				}
				// Remove generated id to be able to compare the widget with the expected one
				got.Widgets[i].Id = ""
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("Wrong template received, diff:\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}
