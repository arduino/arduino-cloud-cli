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

package iotapiraw

type BoardType struct {
	FQBN                 *string  `json:"fqbn,omitempty"`
	Label                string   `json:"label"`
	MinProvSketchVersion *string  `json:"min_provisioning_sketch_version,omitempty"`
	MinWiFiVersion       *string  `json:"min_provisioning_wifi_version,omitempty"`
	Provisioning         *string  `json:"provisioning,omitempty"`
	Tags                 []string `json:"tags"`
	Type                 string   `json:"type"`
	Vendor               string   `json:"vendor"`
	OTAAvailable         *bool    `json:"ota_available,omitempty"`
}

type BoardTypeList []BoardType

type Prov2SketchBinRes struct {
	Binary   string `json:"bin"`
	FileName string `json:"filename"`
	FQBN     string `json:"fqbn"`
	Name     string `json:"name"`
	SHA256   string `json:"sha256"`
}
