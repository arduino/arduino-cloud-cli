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

package serial

var (
	// msgStart is the initial byte sequence of every packet
	msgStart = [2]byte{0x55, 0xAA}
	// msgEnd is the final byte sequence of every packet
	msgEnd = [2]byte{0xAA, 0x55}
)

const (
	// Position of payload field
	payloadField = 5
	// Position of payload length field
	payloadLenField = 3
	// Length of payload length field
	payloadLenFieldLen = 2
	// Length of the signature field
	crcFieldLen = 2
)

// MsgType indicates the type of the packet
type MsgType byte

const (
	None MsgType = iota
	Cmd
	Data
	Response
)

// Command indicates the command that should be
// executed on the board to be provisioned.
type Command byte

const (
	SketchInfo Command = iota + 1
	CSR
	Locked
	GetLocked
	WriteCrypto
	BeginStorage
	SetDeviceID
	SetYear
	SetMonth
	SetDay
	SetHour
	SetValidity
	SetCertSerial
	SetAuthKey
	SetSignature
	EndStorage
	ReconstructCert
)
