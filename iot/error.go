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

package iot

import (
	"encoding/json"
	"fmt"

	iotclient "github.com/arduino/iot-client-go"
)

// errorDetail takes a generic iot-client-go error
// and tries to return a more detailed error.
func errorDetail(err error) error {
	apiErr, ok := err.(iotclient.GenericOpenAPIError)
	if !ok {
		return err
	}

	modErr, ok := apiErr.Model().(iotclient.ModelError)
	if ok {
		return fmt.Errorf("%w: %s", err, modErr.Detail)
	}

	body := make(map[string]interface{})
	if bodyErr := json.Unmarshal(apiErr.Body(), &body); bodyErr != nil {
		return err
	}
	detail, ok := body["detail"]
	if !ok {
		return err
	}
	return fmt.Errorf("%w: %v", err, detail)
}
