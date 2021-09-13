package device

import (
	"testing"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// Test variables
var (
	portsNoBoards = []*rpc.DetectedPort{
		{
			Address: "ACM0",
			Boards:  []*rpc.BoardListItem{},
		},
		{
			Address: "ACM1",
			Boards:  []*rpc.BoardListItem{},
		},
	}

	portsTwoBoards = []*rpc.DetectedPort{
		{
			Address: "ACM0",
			Boards: []*rpc.BoardListItem{
				{Fqbn: "arduino:samd:nano_33_iot"},
			},
		},
		{
			Address: "ACM1",
			Boards: []*rpc.BoardListItem{
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
		want   *board
	}{

		{
			name:   "port-filter",
			filter: &CreateParams{Fqbn: nil, Port: stringPointer("ACM1")},
			ports:  portsTwoBoards,
			want:   &board{fqbn: "arduino:avr:uno", port: "ACM1"},
		},

		{
			name:   "fqbn-filter",
			filter: &CreateParams{Fqbn: stringPointer("arduino:avr:uno"), Port: nil},
			ports:  portsTwoBoards,
			want:   &board{fqbn: "arduino:avr:uno", port: "ACM1"},
		},

		{
			name:   "no-filter-noboards",
			filter: &CreateParams{Fqbn: nil, Port: nil},
			ports:  portsNoBoards,
			want:   nil,
		},

		{
			name:   "no-filter",
			filter: &CreateParams{Fqbn: nil, Port: nil},
			ports:  portsTwoBoards,
			// first board found is selected
			want: &board{fqbn: "arduino:samd:nano_33_iot", port: "ACM0"},
		},

		{
			name:   "both-filter-noboards",
			filter: &CreateParams{Fqbn: stringPointer("arduino:avr:uno"), Port: stringPointer("ACM1")},
			ports:  portsNoBoards,
			want:   nil,
		},

		{
			name:   "both-filter-found",
			filter: &CreateParams{Fqbn: stringPointer("arduino:avr:uno"), Port: stringPointer("ACM1")},
			ports:  portsTwoBoards,
			want:   &board{fqbn: "arduino:avr:uno", port: "ACM1"},
		},

		{
			name:   "both-filter-notfound",
			filter: &CreateParams{Fqbn: stringPointer("arduino:avr:uno"), Port: stringPointer("ACM0")},
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
				t.Errorf("Expected nil board, received not nil board with port %s and fqbn %s", got.port, got.fqbn)

			} else if got == nil && tt.want != nil {
				t.Errorf("Expected not nil board with port %s and fqbn %s, received a nil board", tt.want.port, tt.want.fqbn)

			} else if got.port != tt.want.port || got.fqbn != tt.want.fqbn {
				t.Errorf("Expected board with port %s and fqbn %s, received board with port %s and fqbn %s",
					tt.want.port, tt.want.fqbn, got.port, got.fqbn)
			}
		})
	}
}
