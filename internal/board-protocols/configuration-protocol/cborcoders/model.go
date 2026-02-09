// This file is part of arduino-cloud-cli.
//
// Copyright ARDUINO SRL http://www.arduino.cc/)
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

package cborcoders

import (
	"fmt"
)

// Provisioning commands
type ProvisioningStatusMessage struct {
	_      struct{} `cbor:",toarray"`
	Status int16
}

func (t ProvisioningStatusMessage) String() string {
	return fmt.Sprintf("ProvisioningStatusMessage{Status: %d}", t.Status)
}

type WiFiNetwork struct {
	_    struct{} `cbor:",toarray"`
	SSID string
	RSSI int
}

func (w WiFiNetwork) String() string {
	return fmt.Sprintf("WiFiNetwork{SSID: %s, RSSI: %d}", w.SSID, w.RSSI)
}

type WiFiNetworks []WiFiNetwork

type ProvisioningUniqueIdMessage struct {
	_        struct{} `cbor:",toarray"`
	UniqueId [32]uint8
}

func (t ProvisioningUniqueIdMessage) String() string {
	return fmt.Sprintf("ProvisioningUniqueIdMessage{UniqueId: %v}", t.UniqueId)
}

type ProvisioningSignatureMessage struct {
	_         struct{} `cbor:",toarray"`
	Signature [268]uint8
}

func (t ProvisioningSignatureMessage) String() string {
	return fmt.Sprintf("ProvisioningSignatureMessage{Signature: %v}", t.Signature)
}

type ProvisioningPublicKeyMessage struct {
	_                     struct{} `cbor:",toarray"`
	ProvisioningPublicKey string
}

func (t ProvisioningPublicKeyMessage) String() string {
	return fmt.Sprintf("ProvisioningPublicKeyMessage{ProvisioningPublicKey: %s}", t.ProvisioningPublicKey)
}

type ProvisioningBLEMacAddressMessage struct {
	_             struct{} `cbor:",toarray"`
	BLEMacAddress [6]uint8
}

func (t ProvisioningBLEMacAddressMessage) String() string {
	return fmt.Sprintf("ProvisioningSignatureMessage{Signature: %v}", t.BLEMacAddress)
}

type ProvisioningWiFiFWVersionMessage struct {
	_             struct{} `cbor:",toarray"`
	WiFiFWVersion string
}

func (t ProvisioningWiFiFWVersionMessage) String() string {
	return fmt.Sprintf("ProvisioningWiFiFWVersionMessage{WiFiFWVersion: %s}", t.WiFiFWVersion)
}

type ProvisioningSketchVersionMessage struct {
	_                         struct{} `cbor:",toarray"`
	ProvisioningSketchVersion string
}

func (t ProvisioningSketchVersionMessage) String() string {
	return fmt.Sprintf("ProvisioningSketchVersionMessage{ProvisioningSketchVersion: %s}", t.ProvisioningSketchVersion)
}

type ProvisioningNetworkConfigLibVersionMessage struct {
	_                       struct{} `cbor:",toarray"`
	NetworkConfigLibVersion string
}

func (t ProvisioningNetworkConfigLibVersionMessage) String() string {
	return fmt.Sprintf("ProvisioningNetworkConfigLibVersionMessage{NetworkConfigLibVersion: %s}", t.NetworkConfigLibVersion)
}

type ProvisioningTimestampMessage struct {
	_         struct{} `cbor:",toarray"`
	Timestamp uint64
}

func (t ProvisioningTimestampMessage) String() string {
	return fmt.Sprintf("ProvisioningTimestampMessage{Timestamp: %d}", t.Timestamp)
}

type ProvisioningCommandsMessage struct {
	_       struct{} `cbor:",toarray"`
	Command uint8
}

func (t ProvisioningCommandsMessage) String() string {
	return fmt.Sprintf("ProvisioningCommandsMessage{Command: %d}", t.Command)
}

type ProvisioningWifiConfigMessage struct {
	_    struct{} `cbor:",toarray"`
	SSID string
	PWD  string
}

func (t ProvisioningWifiConfigMessage) String() string {
	return fmt.Sprintf("ProvisioningWifiConfigMessage{SSID: %s, PWD: %s}", t.SSID, t.PWD)
}

type ProvisioningLoRaConfigMessage struct {
	_           struct{} `cbor:",toarray"`
	AppEui      string
	AppKey      string
	Band        uint8
	ChannelMask string
	DeviceClass string
}

func (t ProvisioningLoRaConfigMessage) String() string {
	return fmt.Sprintf("ProvisioningLoRaConfigMessage{appEui: %s, appKey: %s, band: %d, channelMask: %s, deviceClass: %s}", t.AppEui, t.AppKey, t.Band, t.ChannelMask, t.DeviceClass)
}

type ProvisioningCATM1ConfigMessage struct {
	_     struct{} `cbor:",toarray"`
	PIN   string
	Band  []uint32
	Apn   string
	Login string
	Pass  string
}

func (t ProvisioningCATM1ConfigMessage) String() string {
	return fmt.Sprintf("ProvisioningCATM1ConfigMessage{PIN: %s, Band: %v, Apn: %s, Login: %s, Pass: %s}", t.PIN, t.Band, t.Apn, t.Login, t.Pass)
}

type ProvisioningEthernetConfigMessage struct {
	_               struct{} `cbor:",toarray"`
	Static_ip       []uint8  `cbor:",toarray"`
	Dns             []uint8  `cbor:",toarray"`
	Gateway         []uint8  `cbor:",toarray"`
	Netmask         []uint8  `cbor:",toarray"`
	Timeout         uint
	ResponseTimeout uint
}

func (t ProvisioningEthernetConfigMessage) String() string {
	return fmt.Sprintf("ProvisioningEthernetConfigMessage{Static_ip: %v, Dns: %v, Gateway: %v, Netmask: %v, Timeout: %d, ResponseTimeout: %d}", t.Static_ip, t.Dns, t.Gateway, t.Netmask, t.Timeout, t.ResponseTimeout)
}

type ProvisioningCellularConfigMessage struct {
	_     struct{} `cbor:",toarray"`
	PIN   string
	Apn   string
	Login string
	Pass  string
}

func (t ProvisioningCellularConfigMessage) String() string {
	return fmt.Sprintf("ProvisioningCellularConfigMessage{PIN: %s, Apn: %s, Login: %s, Pass: %s}", t.PIN, t.Apn, t.Login, t.Pass)
}

type ProvisioningGSMConfigMessage struct {
	_     struct{} `cbor:",toarray"`
	PIN   string
	Apn   string
	Login string
	Pass  string
}

func (t ProvisioningGSMConfigMessage) String() string {
	return fmt.Sprintf("ProvisioningGSMConfigMessage{PIN: %s, Apn: %s, Login: %s, Pass: %s}", t.PIN, t.Apn, t.Login, t.Pass)
}

type ProvisioningNBConfigMessage struct {
	_     struct{} `cbor:",toarray"`
	PIN   string
	Apn   string
	Login string
	Pass  string
}

func (t ProvisioningNBConfigMessage) String() string {
	return fmt.Sprintf("ProvisioningNBConfigMessage{PIN: %s, Apn: %s, Login: %s, Pass: %s}", t.PIN, t.Apn, t.Login, t.Pass)
}
