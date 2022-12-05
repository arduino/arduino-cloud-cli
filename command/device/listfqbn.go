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
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	arduinoPackage = "arduino"
	esp32Package   = "esp32"
	esp8266Package = "esp8266"
)

// FQBNInfo contains the details of a FQBN.
type FQBNInfo struct {
	Value   string `json:"fqbn"`
	Name    string `json:"name"`
	Package string `json:"package"`
}

// ListFQBN command returns a list of the supported FQBN.
func ListFQBN(ctx context.Context) ([]FQBNInfo, error) {
	url := "https://builder.arduino.cc/v3/boards/"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve boards: %w", err)
	}

	h := &http.Client{Timeout: time.Second * 5}
	resp, err := h.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve boards: %w", err)
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

	return filterFQBN(fqbnList.Items), nil
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
			compatible := append(cryptoFQBN, loraFQBN...)
			for _, b := range compatible {
				if fqbn.Value == b {
					filtered = append(filtered, fqbn)
					break
				}
			}
		}
	}
	return filtered
}
