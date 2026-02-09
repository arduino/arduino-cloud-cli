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

package frame

import (
	"bytes"
	"encoding/binary"

	"github.com/howeyc/crc16"
)

/*
 * The ArduinoBoardConfiguration Protocol frame structure
 * 0x55 0xaa <type> <len> <payload> <crc> 0xaa 0x55
 *  ____________________________________________________________________________________________________________________________________
 * | Byte[0] | Byte[1] | Byte[2] | Byte[3] | Byte[4] | Byte[5].......Byte[len -3] | Byte[len-2] | Byte[len-1] | Byte[len] | Byte[len+1] |
 * |______________________HEADER_____________________|__________ PAYLOAD _________|___________ CRC ___________|________ FOOTER _________|
 * | 0x55    | 0xaa    | <type>  | <len>             | <payload>                  | <crc>                     | 0xaa      | 0x55        |
 * |____________________________________________________________________________________________________________________________________|
 * <type> = MessageType: 2 = DATA, 3 = TRANSMISSION_CONTROL
 * <len> = length of the payload + 2 bytes for the CRC
 * <payload> = the data to be sent or received
 * <crc> = CRC16 of the payload
 */

var (
	// msgStart is the initial byte sequence of every packet.
	msgStart = [2]byte{0x55, 0xAA}
	// msgEnd is the final byte sequence of every packet.
	msgEnd = [2]byte{0xAA, 0x55}
)

const (
	// headerLen indicates the length of the header.
	headerLen = 5
	// payloadField indicates the position of payload field.
	payloadField = 5
	// payloadLenField indicates the position of payload length field.
	payloadLenField = 3
	// payloadLenFieldLen indicatest the length of payload length field.
	payloadLenFieldLen = 2
	// crcFieldLen indicates the length of the signature field.
	crcFieldLen = 2
)

// MsgType indicates the type of the packet.
type MsgType byte

const (
	None MsgType = iota
	Cmd
	Data
	Response
	TransmissionControl = Response // Alias for Response, used for clarity in ArduinoBoardConfiguration Protocol
)

type Frame struct {
	header     []byte
	payload    []byte
	crc        []byte
	footer     []byte
	payloadLen int
	length     int
}

func (p *Frame) FillByte(b byte) bool {

	if len(p.header) < headerLen {
		if p.fillHeader(b) {
			p.extractPayloadLength()
			p.length = headerLen + p.payloadLen + crcFieldLen + len(msgEnd)
		}
	} else if len(p.payload) < p.payloadLen {
		p.payload = append(p.payload, b)
	} else if len(p.crc) < crcFieldLen {
		p.crc = append(p.crc, b)
	} else if len(p.footer) < 2 {
		p.footer = append(p.footer, b)
	}

	return len(p.header) == headerLen && len(p.payload) == p.payloadLen && len(p.crc) == crcFieldLen && len(p.footer) == 2
}

func (p *Frame) fillHeader(data byte) bool {
	p.header = append(p.header, data)
	return len(p.header) == 5
}

func (p *Frame) extractPayloadLength() bool {
	if len(p.header) < 5 {
		return false
	}
	p.payloadLen = int(binary.BigEndian.Uint16(p.header[payloadLenField:])) - crcFieldLen
	return true
}

func (p *Frame) Validate() bool {
	if len(p.header) != headerLen && !bytes.Equal(p.header[:2], msgStart[:]) {
		return false
	}

	if p.payloadLen == 0 {
		return false
	}

	if len(p.payload) != p.payloadLen {
		return false
	}

	if len(p.crc) != crcFieldLen {
		return false
	}

	if len(p.footer) != 2 && !bytes.Equal(p.footer[:], msgEnd[:]) {
		return false
	}

	ch := crc16.Checksum(p.payload, crc16.CCITTTable)
	// crc is contained in the last bytes of the payload
	cp := binary.BigEndian.Uint16(p.crc)
	if ch != cp {
		return false
	}

	return true
}

func (p *Frame) GetPayload() []byte {
	if !p.Validate() {
		return nil
	}
	return p.payload
}

func (p *Frame) GetType() MsgType {
	if !p.Validate() {
		return None
	}
	return MsgType(p.header[2])
}

func (p *Frame) GetLength() int {
	return p.length
}

func (p *Frame) ToBytes() []byte {
	return append(append(append(p.header, p.payload...), p.crc...), p.footer...)
}

func (p *Frame) SetPayload(payload []byte) {
	p.length += len(payload)
	p.payload = payload
	p.payloadLen = len(payload)
}

func (p *Frame) SetHeader(header []byte) {
	p.length += len(header)
	p.header = header
}

func (p *Frame) SetCrc(crc []byte) {
	p.length += len(crc)
	p.crc = crc
}

func (p *Frame) SetFooter(footer []byte) {
	p.length += len(footer)
	p.footer = footer
}

func CreateFrame(data []byte, mType MsgType) Frame {
	// Create the packet
	packet := Frame{}
	packetHeader := append(msgStart[:], byte(mType))

	// Append the packet length
	bLen := make([]byte, payloadLenFieldLen)
	binary.BigEndian.PutUint16(bLen, (uint16(len(data) + crcFieldLen)))
	packetHeader = append(packetHeader, bLen...)

	packet.SetHeader(packetHeader)
	// Append the message payload
	packet.SetPayload(data)

	// Calculate and append the message signature
	ch := crc16.Checksum(data, crc16.CCITTTable)
	checksum := make([]byte, crcFieldLen)
	binary.BigEndian.PutUint16(checksum, ch)
	packet.SetCrc(checksum)

	// Append final byte sequence
	packet.SetFooter(msgEnd[:])
	return packet
}
