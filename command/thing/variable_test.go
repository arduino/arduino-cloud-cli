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

package thing

import (
	"testing"

	iotclient "github.com/arduino/iot-client-go/v3"
	"github.com/stretchr/testify/assert"
)

func validParams() *CreateVariableParams {
	return &CreateVariableParams{
		ThingID:        "thing-id",
		Name:           "temperature",
		Type:           "float",
		Permission:     "read_only",
		UpdateStrategy: "on_change",
	}
}

func TestBuildProperty(t *testing.T) {
	property, err := buildProperty(validParams())
	assert.NoError(t, err)
	assert.Equal(t, &iotclient.Property{
		Name:           "temperature",
		Type:           "FLOAT",
		Permission:     "READ_ONLY",
		UpdateStrategy: "ON_CHANGE",
	}, property)
}

func TestBuildPropertyTimed(t *testing.T) {
	params := validParams()
	params.UpdateStrategy = "TIMED"
	params.UpdateParameter = float64Ptr(10)
	varName := "temp_c"
	params.VariableName = &varName

	property, err := buildProperty(params)
	assert.NoError(t, err)
	assert.Equal(t, "TIMED", property.UpdateStrategy)
	assert.Equal(t, 10.0, *property.UpdateParameter)
	assert.Equal(t, "temp_c", *property.VariableName)
}

func TestBuildPropertyErrors(t *testing.T) {
	tests := []struct {
		name   string
		modify func(p *CreateVariableParams)
		errMsg string
	}{
		{"missing thing id", func(p *CreateVariableParams) { p.ThingID = "" }, "thing ID not specified"},
		{"missing name", func(p *CreateVariableParams) { p.Name = "" }, "variable name not specified"},
		{"missing type", func(p *CreateVariableParams) { p.Type = "" }, "variable type not specified"},
		{"invalid type", func(p *CreateVariableParams) { p.Type = "DOUBLE" }, "invalid variable type"},
		{"invalid variable name", func(p *CreateVariableParams) {
			v := "1temp"
			p.VariableName = &v
		}, "invalid variable name"},
		{"invalid permission", func(p *CreateVariableParams) { p.Permission = "WRITE_ONLY" }, "invalid permission"},
		{"invalid update strategy", func(p *CreateVariableParams) { p.UpdateStrategy = "ALWAYS" }, "invalid update strategy"},
		{"timed without parameter", func(p *CreateVariableParams) { p.UpdateStrategy = "TIMED" }, "requires a positive update parameter"},
		{"negative update parameter", func(p *CreateVariableParams) { p.UpdateParameter = float64Ptr(-1) }, "cannot be negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := validParams()
			tt.modify(params)
			_, err := buildProperty(params)
			assert.ErrorContains(t, err, tt.errMsg)
		})
	}
}

func TestGetVariableInfo(t *testing.T) {
	varName := "temperature"
	info := getVariableInfo(&iotclient.ArduinoProperty{
		Id:              "prop-id",
		Name:            "Temperature",
		VariableName:    &varName,
		Type:            "FLOAT",
		Permission:      "READ_ONLY",
		UpdateStrategy:  "TIMED",
		UpdateParameter: float64Ptr(5),
	})
	assert.Equal(t, &VariableInfo{
		ID:              "prop-id",
		Name:            "Temperature",
		VariableName:    "temperature",
		Type:            "FLOAT",
		Permission:      "READ_ONLY",
		UpdateStrategy:  "TIMED",
		UpdateParameter: float64Ptr(5),
	}, info)
}

func float64Ptr(f float64) *float64 {
	return &f
}
