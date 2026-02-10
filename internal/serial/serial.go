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

package serial

import (
	"errors"
	"fmt"
	"time"

	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/frame"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/transport"
	"go.bug.st/serial"
)

// Serial is a wrapper of serial port interface that
// features specific functions to send provisioning
// commands through the serial port to an arduino device.
type Serial struct {
	port      serial.Port
	connected bool
}

// NewSerial instantiate and returns a Serial instance.
// The Serial Connect method should be called before using
// its send/receive functions.
func NewSerial() *Serial {
	s := &Serial{}
	s.connected = false
	return s
}

// Connect tries to connect Serial to a specific serial port.
func (s *Serial) Connect(params transport.TransportInterfaceParams) error {
	mode := &serial.Mode{
		BaudRate: params.BoundRate,
	}
	port, err := serial.Open(params.Port, mode)
	if err != nil {
		err = fmt.Errorf("%s: %w", "connecting to serial port", err)
		return err
	}
	s.port = port
	s.connected = true
	s.port.SetReadTimeout(time.Millisecond * 2500)
	return nil
}

func (s *Serial) Send(data []byte) error {
	_, err := s.port.Write(data)
	if err != nil {
		err = fmt.Errorf("%s: %w", "sending message through serial", err)
		return err
	}

	return nil
}

// Close should be used when the Serial connection isn't used anymore.
// After that, Serial could Connect again to any port.
func (s *Serial) Close() error {
	s.connected = false
	return s.port.Close()
}

func (s *Serial) Receive(timeoutSeconds int) ([]frame.Frame, error) {
	if !s.connected {
		return nil, errors.New("serial port not connected")
	}

	expireTimeout := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)
	received := false
	transportController := transport.NewTransportController()
	packets := []frame.Frame{}

	for !received && time.Now().Before(expireTimeout) {
		buffer := make([]byte, 1024)
		n, err := s.port.Read(buffer)
		if err != nil {
			return nil, err
		}

		packets = transportController.HandleReceivedData(buffer[:n])
		if len(packets) > 0 {
			received = true
		}
	}

	if !received {
		return nil, fmt.Errorf("no response received after %d seconds", timeoutSeconds)
	}

	return packets, nil
}

func (s *Serial) Type() transport.InterfaceType {
	return transport.Serial
}

func (s *Serial) Connected() bool {
	return s.connected
}
