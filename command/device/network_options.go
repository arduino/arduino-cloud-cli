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

package device

type WiFiSetting struct {
	SSID string `json:"ssid"` // Max length of ssid is 32 + \0
	PWD  string `json:"pwd"`  // Max length of password is 63 + \0
}

type IPAddr struct {
	Type  uint8    `json:"type"`
	Bytes [16]byte `json:"bytes"`
}

type EthernetSetting struct {
	IP              IPAddr `json:"ip"`
	DNS             IPAddr `json:"dns"`
	Gateway         IPAddr `json:"gateway"`
	Netmask         IPAddr `json:"netmask"`
	Timeout         uint   `json:"timeout"`
	ResponseTimeout uint   `json:"response_timeout"`
}

type CellularSetting struct {
	PIN   string `json:"pin"`   // Max length of pin is 8 + \0
	APN   string `json:"apn"`   // Max length of apn is 100 + \0
	Login string `json:"login"` // Max length of login is 32 + \0
	Pass  string `json:"pass"`  // Max length of pass is 32 + \0
}

type CATM1Setting struct {
	PIN   string    `json:"pin"`   // Max length of pin is 8 + \0
	APN   string    `json:"apn"`   // Max length of apn is 100 + \0
	Login string    `json:"login"` // Max length of login is 32 + \0
	Pass  string    `json:"pass"`  // Max length of pass is 32 + \0
	Band  [4]uint32 `json:"band"`
}

type LoraSetting struct {
	AppEUI      string `json:"appeui"` // appeui is 8 octets * 2 (hex format) + \0
	AppKey      string `json:"appkey"` // appeui is 16 octets * 2 (hex format) + \0
	Band        uint8  `json:"band"`
	ChannelMask string `json:"channel_mask"`
	DeviceClass string `json:"device_class"`
}

type NetConfig struct {
	Type            int32           `json:"type"`
	WiFi            WiFiSetting     `json:"wifi,omitempty"`
	Eth             EthernetSetting `json:"eth,omitempty"`
	NB              CellularSetting `json:"nb,omitempty"`
	GSM             CellularSetting `json:"gsm,omitempty"`
	CATM1           CATM1Setting    `json:"catm1,omitempty"`
	CellularSetting CellularSetting `json:"cellular,omitempty"`
	Lora            LoraSetting     `json:"lora,omitempty"`
}
