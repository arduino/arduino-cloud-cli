// This file is part of arduino-cloud-cli.
//
// Copyright (C) 2025 ARDUINO SA (http://www.arduino.cc/)
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

package provisioningapi

type RegisterBoardData struct {
	PID              string  `json:"pid"`
	PublicKey        string  `json:"public_key"`
	Serial           *string `json:"serial"`
	UniqueHardwareID string  `json:"unique_hardware_id"`
	VID              string  `json:"vid"`
}

type ClaimData struct {
	BLEMac         string `json:"ble_mac"`
	BoardToken     string `json:"board_token"`
	ConnectionType string `json:"connection_type"`
	DeviceName     string `json:"device_name"`
}

type BadResponse struct {
	Err     string `json:"err"`
	ErrCode int    `json:"err_code"`
}

type ClaimResponse struct {
	BoardId string `json:"id"`
}

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

type Onboarding struct {
	ID               string  `json:"id"`
	UniqueHardwareID string  `json:"unique_hardware_id"`
	DeviceName       string  `json:"device_name"`
	ConnectionType   string  `json:"connection_type"`
	DeviceID         *string `json:"device_id"`
	UserID           string  `json:"user_id"`
	OrgID            *string `json:"org_id"`
	BLEMac           string  `json:"ble_mac"`
	CreatedAt        string  `json:"created_at"`
	ProvisionedAt    *string `json:"provisioned_at"`
	ClaimedAt        string  `json:"claimed_at"`
	EndedAt          *string `json:"ended_at"`
	FQBN             string  `json:"fqbn"`
}

type OnboardingsResponse struct {
	Onboardings []Onboarding `json:"onboardings"`
}
