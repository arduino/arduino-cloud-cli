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
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	iotclient "github.com/arduino/iot-client-go/v3"
)

var variableNameRegexp = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// variableTypes contains the variable types accepted by Arduino IoT Cloud,
// the same values used in the 'type' field of thing template variables.
var variableTypes = map[string]struct{}{
	"ANALOG": {}, "CHARSTRING": {}, "FLOAT": {}, "INT": {}, "LENGHT_C": {}, "LENGHT_I": {},
	"LENGHT_M": {}, "PERCENTAGE": {}, "STATUS": {}, "TEMPERATURE_C": {}, "TEMPERATURE_F": {},
	"METER": {}, "KILOGRAM": {}, "GRAM": {}, "SECOND": {}, "AMPERE": {}, "KELVIN": {}, "CANDELA": {},
	"MOLE": {}, "HERTZ": {}, "RADIAN": {}, "STERADIAN": {}, "NEWTON": {}, "PASCAL": {}, "JOULE": {},
	"WATT": {}, "COULOMB": {}, "VOLT": {}, "FARAD": {}, "OHM": {}, "SIEMENS": {}, "WEBER": {},
	"TESLA": {}, "HENRY": {}, "DEGREES_CELSIUS": {}, "LUMEN": {}, "LUX": {}, "BECQUEREL": {},
	"GRAY": {}, "SIEVERT": {}, "KATAL": {}, "SQUARE_METER": {}, "CUBIC_METER": {}, "LITER": {},
	"METER_PER_SECOND": {}, "METER_PER_SQUARE_SECOND": {}, "CUBIC_METER_PER_SECOND": {},
	"LITER_PER_SECOND": {}, "WATT_PER_SQUARE_METER": {}, "CANDELA_PER_SQUARE_METER": {}, "BIT": {},
	"BIT_PER_SECOND": {}, "DEGREES_LATITUDE": {}, "DEGREES_LONGITUDE": {}, "PH_VALUE": {},
	"DECIBEL": {}, "DECIBEL_1W": {}, "BEL": {}, "COUNT": {}, "RATIO_DIV": {}, "RATIO_MOD": {},
	"PERCENTAGE_RELATIVE_HUMIDITY": {}, "PERCENTAGE_BATTERY_LEVEL": {}, "SECONDS_BATTERY_LEVEL": {},
	"EVENT_RATE_SECOND": {}, "EVENT_RATE_MINUTE": {}, "HEART_RATE": {}, "HEART_BEATS": {},
	"SIEMENS_PER_METER": {}, "LOCATION": {}, "COLOR_HSB": {}, "COLOR_RGB": {},
	"GENERIC_COMPLEX_PROPERTY": {}, "HOME_COLORED_LIGHT": {}, "HOME_DIMMED_LIGHT": {},
	"HOME_LIGHT": {}, "HOME_CONTACT_SENSOR": {}, "HOME_MOTION_SENSOR": {}, "HOME_SMART_PLUG": {},
	"HOME_TEMPERATURE": {}, "HOME_TEMPERATURE_C": {}, "HOME_TEMPERATURE_F": {}, "HOME_SWITCH": {},
	"HOME_TELEVISION": {}, "GOOGLE_HOME_COLORED_LIGHT": {}, "GOOGLE_HOME_DIMMED_LIGHT": {},
	"GOOGLE_HOME_LIGHT": {}, "GOOGLE_HOME_CONTACT_SENSOR": {}, "GOOGLE_HOME_MOTION_SENSOR": {},
	"GOOGLE_HOME_SMART_PLUG": {}, "GOOGLE_HOME_TEMPERATURE": {}, "GOOGLE_HOME_TEMPERATURE_C": {},
	"GOOGLE_HOME_TEMPERATURE_F": {}, "GOOGLE_HOME_SWITCH": {}, "SMART_HOME_COLORED_LIGHT": {},
	"SMART_HOME_DIMMED_LIGHT": {}, "SMART_HOME_LIGHT": {}, "SMART_HOME_CONTACT_SENSOR": {},
	"SMART_HOME_MOTION_SENSOR": {}, "SMART_HOME_SMART_PLUG": {}, "SMART_HOME_TEMPERATURE": {},
	"SMART_HOME_TEMPERATURE_C": {}, "SMART_HOME_TEMPERATURE_F": {}, "SMART_HOME_SWITCH": {},
	"SMART_HOME_TELEVISION": {}, "ENERGY": {}, "FORCE": {}, "TEMPERATURE": {}, "POWER": {},
	"ELECTRIC_CURRENT": {}, "ELECTRIC_POTENTIAL": {}, "ELECTRICAL_RESISTANCE": {}, "CAPACITANCE": {},
	"TIME": {}, "FREQUENCY": {}, "DATA_RATE": {}, "ACCELERATION": {}, "AREA": {}, "LENGTH": {},
	"VELOCITY": {}, "MASS": {}, "VOLUME": {}, "FLOW_RATE": {}, "ANGLE": {}, "ILLUMINANCE": {},
	"LUMINOUS_FLUX": {}, "LUMINANCE": {}, "LUMINOUS_INTENSITY": {}, "LOGARITHMIC_QUANTITY": {},
	"PRESSURE": {}, "INFORMATION_CONTENT": {}, "SCHEDULE": {},
}

// VariableInfo contains the main parameters of
// a variable of an Arduino IoT Cloud thing.
type VariableInfo struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	VariableName    string   `json:"variable_name"`
	Type            string   `json:"type"`
	Permission      string   `json:"permission"`
	UpdateStrategy  string   `json:"update_strategy"`
	UpdateParameter *float64 `json:"update_parameter,omitempty"`
}

// CreateVariableParams contains the parameters needed
// to add a new variable to an existing thing.
type CreateVariableParams struct {
	ThingID         string
	Name            string
	VariableName    *string  // Optional: sketch variable name
	Type            string   // Variable type, as used in thing templates (e.g. FLOAT, INT, HOME_SWITCH)
	Permission      string   // READ_ONLY or READ_WRITE
	UpdateStrategy  string   // ON_CHANGE or TIMED
	UpdateParameter *float64 // Optional: change threshold for ON_CHANGE, period in seconds for TIMED
}

// CreateVariable adds a new variable to an existing thing,
// preserving the thing ID and everything bound to it.
func CreateVariable(ctx context.Context, params *CreateVariableParams, cred *config.Credentials) (*VariableInfo, error) {
	property, err := buildProperty(params)
	if err != nil {
		return nil, err
	}

	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return nil, err
	}

	newProperty, err := iotClient.PropertyCreate(ctx, params.ThingID, property)
	if err != nil {
		return nil, err
	}
	return getVariableInfo(newProperty), nil
}

// ListVariables returns the variables of a thing.
func ListVariables(ctx context.Context, thingID string, cred *config.Credentials) ([]VariableInfo, error) {
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return nil, err
	}

	properties, err := iotClient.PropertyList(ctx, thingID)
	if err != nil {
		return nil, err
	}

	vars := make([]VariableInfo, 0, len(properties))
	for i := range properties {
		vars = append(vars, *getVariableInfo(&properties[i]))
	}
	return vars, nil
}

// DeleteVariable deletes a variable from a thing.
func DeleteVariable(ctx context.Context, thingID, variableID string, cred *config.Credentials) error {
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return err
	}
	return iotClient.PropertyDelete(ctx, thingID, variableID)
}

func buildProperty(params *CreateVariableParams) (*iotclient.Property, error) {
	if params.ThingID == "" {
		return nil, errors.New("thing ID not specified")
	}
	if params.Name == "" {
		return nil, errors.New("variable name not specified")
	}
	if params.VariableName != nil && !variableNameRegexp.MatchString(*params.VariableName) {
		return nil, fmt.Errorf("invalid variable name %q: it must be a valid sketch identifier", *params.VariableName)
	}

	varType := strings.ToUpper(params.Type)
	if varType == "" {
		return nil, errors.New("variable type not specified")
	}
	if _, ok := variableTypes[varType]; !ok {
		return nil, fmt.Errorf("invalid variable type %q: use one of the types accepted in thing templates (e.g. FLOAT, INT, STATUS, CHARSTRING)", params.Type)
	}

	permission := strings.ToUpper(params.Permission)
	if permission != "READ_ONLY" && permission != "READ_WRITE" {
		return nil, fmt.Errorf("invalid permission %q: allowed values are READ_ONLY, READ_WRITE", params.Permission)
	}

	strategy := strings.ToUpper(params.UpdateStrategy)
	switch strategy {
	case "ON_CHANGE":
	case "TIMED":
		if params.UpdateParameter == nil || *params.UpdateParameter <= 0 {
			return nil, errors.New("update strategy TIMED requires a positive update parameter (period in seconds)")
		}
	default:
		return nil, fmt.Errorf("invalid update strategy %q: allowed values are ON_CHANGE, TIMED", params.UpdateStrategy)
	}
	if params.UpdateParameter != nil && *params.UpdateParameter < 0 {
		return nil, errors.New("update parameter cannot be negative")
	}

	return &iotclient.Property{
		Name:            params.Name,
		VariableName:    params.VariableName,
		Type:            varType,
		Permission:      permission,
		UpdateStrategy:  strategy,
		UpdateParameter: params.UpdateParameter,
	}, nil
}

func getVariableInfo(property *iotclient.ArduinoProperty) *VariableInfo {
	return &VariableInfo{
		ID:              property.Id,
		Name:            property.Name,
		VariableName:    dereferenceString(property.VariableName),
		Type:            property.Type,
		Permission:      property.Permission,
		UpdateStrategy:  property.UpdateStrategy,
		UpdateParameter: property.UpdateParameter,
	}
}
