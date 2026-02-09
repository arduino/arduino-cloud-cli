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

package transport

import (
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/frame"
)

type TransportController struct {
	packetList          []frame.Frame
	receivingPacket     frame.Frame
	filledPacket        bool
	foundStart          bool
	foundFirstByteStart bool
}

func NewTransportController() *TransportController {
	return &TransportController{
		packetList:          make([]frame.Frame, 0),
		receivingPacket:     frame.Frame{},
		filledPacket:        false,
		foundStart:          false,
		foundFirstByteStart: false,
	}
}

func (tc *TransportController) HandleReceivedData(data []byte) []frame.Frame {
	// if in the previous iteration the last byte was the beginning of a new packet,
	// check if the first byte of the current iteration is the second byte of the begin frame
	if tc.foundFirstByteStart {
		if len(data) > 0 && data[0] == 0xaa {
			tc.foundStart = true
			//force fill byte
			tc.filledPacket = tc.receivingPacket.FillByte(0x55)
		}
		tc.foundFirstByteStart = false
	}

	n := tc.searchStartPacket(data)
	if n != -1 && !tc.foundStart {
		tc.foundStart = true
		data = data[n:]
	}

	if tc.foundStart {
		for i := 0; i < len(data); i++ {
			if tc.filledPacket {
				tc.foundStart = false
				n = tc.searchStartPacket(data[i:])
				if n != -1 {
					tc.foundStart = true
					tc.filledPacket = false
					tc.packetList = append(tc.packetList, tc.receivingPacket)
					tc.receivingPacket = frame.Frame{}

					i = i + n
				} else {
					break
				}
			}
			tc.filledPacket = tc.receivingPacket.FillByte(data[i])
		}
	} else {
		// Discard data
		//Check if the last byte is the beginning of a new packet in case the begin is split in two packets
		if len(data) > 0 && data[len(data)-1] == 0x55 {
			tc.foundFirstByteStart = true
		}
	}

	if !tc.filledPacket {
		return nil
	}

	tc.packetList = append(tc.packetList, tc.receivingPacket)
	return tc.packetList
}

func (tc *TransportController) searchStartPacket(data []byte) int {
	for i := 0; i < len(data)-1; i++ {
		if data[i] == 0x55 && data[i+1] == 0xaa {
			return i
		}
	}
	return -1
}
