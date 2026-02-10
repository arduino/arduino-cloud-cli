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

import (
	"strings"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

var (
	cryptoFQBN = []string{
		"arduino:samd:nano_33_iot",
		"arduino:samd:mkrwifi1010",
		"arduino:mbed_nano:nanorp2040connect",
		"arduino:mbed_portenta:envie_m7",
		"arduino:mbed_nicla:nicla_vision",
		"arduino:samd:mkr1000",
		"arduino:samd:mkrgsm1400",
		"arduino:samd:mkrnb1500",
		"arduino:mbed_opta:opta",
		"arduino:mbed_giga:giga",
		"arduino:renesas_uno:unor4wifi",
		"arduino:renesas_portenta:portenta_c33",
	}
	loraFQBN = []string{
		"arduino:samd:mkrwan1310",
		"arduino:samd:mkrwan1300",
	}
)

// board contains details of a physical arduino board.
type board struct {
	fqbn     string
	serial   string
	dType    string
	address  string
	protocol string
}

// isCrypto checks if the board is a valid arduino board with a
// supported crypto-chip.
func (b *board) isCrypto() bool {
	for _, f := range cryptoFQBN {
		if b.fqbn == f {
			return true
		}
	}
	return false
}

// isLora checks if the board is a valid LoRa arduino board.
func (b *board) isLora() bool {
	for _, f := range loraFQBN {
		if b.fqbn == f {
			return true
		}
	}
	return false
}

// boardFromPorts returns a board that matches all the criteria
// passed in. If no criteria are passed, it returns the first board found.
func boardFromPorts(ports []*rpc.DetectedPort, params *CreateParams) *board {
	for _, port := range ports {
		if portFilter(port, params) {
			continue
		}
		boardFound := boardFilter(port.MatchingBoards, params)
		if boardFound != nil {
			b := &board{
				fqbn:     boardFound.Fqbn,
				serial:   port.Port.Properties["serialNumber"],
				dType:    strings.Split(boardFound.Fqbn, ":")[2],
				address:  port.Port.Address,
				protocol: port.Port.Protocol,
			}
			return b
		}
	}

	return nil
}

// portFilter filters out the given port in the following cases:
// - if the port parameter does not match the actual port address.
// - if the the detected port does not contain any board.
// It returns:
// true -> to skip the port.
// false -> to keep the port.
func portFilter(port *rpc.DetectedPort, params *CreateParams) bool {
	if len(port.MatchingBoards) == 0 {
		return true
	}
	if params.Port != nil && *params.Port != port.Port.Address {
		return true
	}
	return false
}

// boardFilter looks for a board which has the same fqbn passed as parameter.
// If fqbn parameter is nil, then the first board found is returned.
// It returns:
// - a board if it is found.
// - nil if no board matching the fqbn parameter is found.
func boardFilter(boards []*rpc.BoardListItem, params *CreateParams) (board *rpc.BoardListItem) {
	if params.FQBN == nil {
		return boards[0]
	}
	for _, b := range boards {
		if b.Fqbn == *params.FQBN {
			return b
		}
	}
	return
}
