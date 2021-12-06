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

package device

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	arduinoPackage = "arduino"
	esp32Package   = "esp32"
	esp8266Package = "esp8266"
)

var (
	// this is temporary... it will be removed when
	// https://github.com/arduino/arduino-cloud-cli/pull/74/files#diff-d891696d5c17ea0eecc6b1c23802cbaf553379e701c5e0e1ff23ee0d26d2877cR27-R39
	// will be merged
	compatibleArduinoFQBN = []string{
		"arduino:samd:nano_33_iot",
		"arduino:samd:mkrwifi1010",
		"arduino:mbed_nano:nanorp2040connect",
		"arduino:mbed_portenta:envie_m7",
		"arduino:samd:mkr1000",
		"arduino:samd:mkrgsm1400",
		"arduino:samd:mkrnb1500",
		"arduino:samd:mkrwan1310",
		"arduino:samd:mkrwan1300",
	}
)

// FQBNInfo contains the details of a FQBN.
type FQBNInfo struct {
	Value   string `json:"fqbn"`
	Name    string `json:"name"`
	Package string `json:"package"`
}

// ListFQBN command returns a list of the supported FQBN.
func ListFQBN() ([]FQBNInfo, error) {
	resp, err := http.Get("https://builder.arduino.cc/v3/boards/")
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve boards from builder.arduino.cc: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading boards from builder.arduino.cc: cannot read response's body: %w", err)
	}

	var fqbnList struct {
		Items []FQBNInfo `json:"items"`
	}
	if err = json.Unmarshal(body, &fqbnList); err != nil {
		return nil, fmt.Errorf("cannot parse boards retrieved from builder.arduino.cc: %w", err)
	}

	fqbnList.Items = filterFQBN(fqbnList.Items)
	return fqbnList.Items, nil
}

// filterFQBN takes a list of fqbn and returns only the
// ones supported by iot cloud.
func filterFQBN(ls []FQBNInfo) []FQBNInfo {
	filtered := make([]FQBNInfo, 0, len(ls))
	for _, fqbn := range ls {
		switch fqbn.Package {

		case esp32Package, esp8266Package:
			filtered = append(filtered, fqbn)

		case arduinoPackage:
			for _, b := range compatibleArduinoFQBN {
				if fqbn.Value == b {
					filtered = append(filtered, fqbn)
					break
				}
			}
		}
	}
	return filtered
}
