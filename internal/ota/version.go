// This file is part of arduino-cloud-cli.
//
// Copyright (C) ARDUINO SRL (http://www.arduino.cc/)
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

package ota

const compressionEnabledMagicNumber = 0x40

// Version contains all the OTA header information
// Check out https://arduino.atlassian.net/wiki/spaces/RFC/pages/1616871540/OTA+header+structure for more
// information on the OTA header specs.
type Version struct {
	HeaderVersion   uint8
	Compression     bool
	Signature       bool
	Spare           uint8
	PayloadTarget   uint8
	PayloadMayor    uint8
	PayloadMinor    uint8
	PayloadPatch    uint8
	PayloadBuildNum uint32
}

// Bytes builds a 8 byte length representation of the Version Struct for the OTA update.
func (v *Version) Bytes() []byte {
	version := []byte{0, 0, 0, 0, 0, 0, 0, 0}

	// Set compression
	if v.Compression {
		version[7] = 0x40
	}

	// Other field are currently not implemented ¯\_(ツ)_/¯

	return version
}

// Bytes builds a 8 byte length representation of the Version Struct for the OTA update.
func decodeVersion(version []byte) Version {

	compressed := (version[7] == compressionEnabledMagicNumber)

	// Other field are currently not implemented ¯\_(ツ)_/¯

	return Version{
		Compression: compressed,
	}
}
