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

// FrequencyPlanInfo describes a LoRa frequency plan.
type FQBNInfo struct {
	Name string `json:"name"`
	FQBN string `json:"fqbn"`
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

	var flist struct {
		FQBN []FQBNInfo `json:"items"`
	}
	if err = json.Unmarshal(body, &flist); err != nil {
		return nil, fmt.Errorf("cannot parse boards retrieved from builder.arduino.cc: %w", err)
	}
	return flist.FQBN, nil
}
