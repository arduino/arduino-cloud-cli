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
	"testing"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// Test variables
var (
	portsNoBoards = []*rpc.DetectedPort{
		{
			Port:           &rpc.Port{Address: "ACM0"},
			MatchingBoards: []*rpc.BoardListItem{},
		},
		{
			Port:           &rpc.Port{Address: "ACM1"},
			MatchingBoards: []*rpc.BoardListItem{},
		},
	}

	portsTwoBoards = []*rpc.DetectedPort{
		{
			Port: &rpc.Port{Address: "ACM0"},
			MatchingBoards: []*rpc.BoardListItem{
				{Fqbn: "arduino:samd:nano_33_iot"},
			},
		},
		{
			Port: &rpc.Port{Address: "ACM1"},
			MatchingBoards: []*rpc.BoardListItem{
				{Fqbn: "arduino:avr:uno"},
			},
		},
	}
)

func stringPointer(s string) *string {
	return &s
}

func TestBoardFromPorts(t *testing.T) {
	tests := []struct {
		name   string
		filter *CreateParams
		ports  []*rpc.DetectedPort
		want   *Board
	}{

		{
			name:   "port-filter",
			filter: &CreateParams{FQBN: nil, Port: stringPointer("ACM1")},
			ports:  portsTwoBoards,
			want:   &Board{Fqbn: "arduino:avr:uno", Address: "ACM1"},
		},

		{
			name:   "fqbn-filter",
			filter: &CreateParams{FQBN: stringPointer("arduino:avr:uno"), Port: nil},
			ports:  portsTwoBoards,
			want:   &Board{Fqbn: "arduino:avr:uno", Address: "ACM1"},
		},

		{
			name:   "no-filter-noboards",
			filter: &CreateParams{FQBN: nil, Port: nil},
			ports:  portsNoBoards,
			want:   nil,
		},

		{
			name:   "no-filter",
			filter: &CreateParams{FQBN: nil, Port: nil},
			ports:  portsTwoBoards,
			// first board found is selected
			want: &Board{Fqbn: "arduino:samd:nano_33_iot", Address: "ACM0"},
		},

		{
			name:   "both-filter-noboards",
			filter: &CreateParams{FQBN: stringPointer("arduino:avr:uno"), Port: stringPointer("ACM1")},
			ports:  portsNoBoards,
			want:   nil,
		},

		{
			name:   "both-filter-found",
			filter: &CreateParams{FQBN: stringPointer("arduino:avr:uno"), Port: stringPointer("ACM1")},
			ports:  portsTwoBoards,
			want:   &Board{Fqbn: "arduino:avr:uno", Address: "ACM1"},
		},

		{
			name:   "both-filter-notfound",
			filter: &CreateParams{FQBN: stringPointer("arduino:avr:uno"), Port: stringPointer("ACM0")},
			ports:  portsTwoBoards,
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := boardFromPorts(tt.ports, tt.filter)

			if got == nil && tt.want == nil {
				return

			} else if got != nil && tt.want == nil {
				t.Errorf("Expected nil board, received not nil board with port %s and fqbn %s", got.Address, got.Fqbn)

			} else if got == nil && tt.want != nil {
				t.Errorf("Expected not nil board with port %s and fqbn %s, received a nil board", tt.want.Address, tt.want.Fqbn)

			} else if got.Address != tt.want.Address || got.Fqbn != tt.want.Fqbn {
				t.Errorf("Expected board with port %s and fqbn %s, received board with port %s and fqbn %s",
					tt.want.Address, tt.want.Fqbn, got.Address, got.Fqbn)
			}
		})
	}
}
