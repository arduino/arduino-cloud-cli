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

type TransportInterfaceParams struct {
	Port      string
	BoundRate int
}

// MsgType indicates the type of the packet.
type InterfaceType byte

const (
	Serial InterfaceType = iota
	BLE
)

type TransportInterface interface {
	Connect(params TransportInterfaceParams) error
	Send(data []byte) error
	Receive(timeoutSeconds int) ([]frame.Frame, error)
	Connected() bool
	Type() InterfaceType
	Close() error
}
