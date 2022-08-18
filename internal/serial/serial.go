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

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"github.com/howeyc/crc16"
	"go.bug.st/serial"
)

// Serial is a wrapper of serial port interface that
// features specific functions to send provisioning
// commands through the serial port to an arduino device.
type Serial struct {
	port serial.Port
}

// NewSerial instantiate and returns a Serial instance.
// The Serial Connect method should be called before using
// its send/receive functions.
func NewSerial() *Serial {
	s := &Serial{}
	return s
}

// Connect tries to connect Serial to a specific serial port.
func (s *Serial) Connect(address string) error {
	mode := &serial.Mode{
		BaudRate: 57600,
	}
	port, err := serial.Open(address, mode)
	if err != nil {
		err = fmt.Errorf("%s: %w", "connecting to serial port", err)
		return err
	}
	s.port = port

	s.port.SetReadTimeout(time.Millisecond * 2500)
	return nil
}

// Send allows to send a provisioning command to a connected arduino device.
func (s *Serial) Send(ctx context.Context, cmd Command, payload []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	payload = append([]byte{byte(cmd)}, payload...)
	msg := encode(Cmd, payload)

	_, err := s.port.Write(msg)
	if err != nil {
		err = fmt.Errorf("%s: %w", "sending message through serial", err)
		return err
	}

	return nil
}

// SendReceive allows to send a provisioning command to a connected arduino device.
// Then, it waits for a response from the device and, if any, returns it.
// If no response is received after 2 seconds, an error is returned.
func (s *Serial) SendReceive(ctx context.Context, cmd Command, payload []byte) ([]byte, error) {
	err := s.Send(ctx, cmd, payload)
	if err != nil {
		return nil, err
	}
	return s.receive(ctx)
}

// Close should be used when the Serial connection isn't used anymore.
// After that, Serial could Connect again to any port.
func (s *Serial) Close() error {
	return s.port.Close()
}

// receive allows to wait for a response from an arduino device under provisioning.
// Its timeout is set to 2 seconds. It returns an error if the response is not valid
// or if the timeout expires.
// TODO: consider refactoring using a more explicit procedure:
// start := s.Read(buff, MsgStartLength)
// payloadLen := s.Read(buff, payloadFieldLen)
func (s *Serial) receive(ctx context.Context) ([]byte, error) {
	buff := make([]byte, 1000)
	var resp []byte

	received := 0
	payloadLen := 0
	// Wait to receive the entire packet that is long as the preamble (from msgStart to payload length field)
	// plus the actual payload length plus the length of the ending sequence.
	for received < (payloadLenField+payloadLenFieldLen)+payloadLen+len(msgEnd) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		n, err := s.port.Read(buff)
		if err != nil {
			err = fmt.Errorf("%s: %w", "receiving from serial", err)
			return nil, err
		}
		if n == 0 {
			break
		}
		received += n
		resp = append(resp, buff[:n]...)

		// Update the payload length as soon as it is received.
		if payloadLen == 0 && received >= (payloadLenField+payloadLenFieldLen) {
			payloadLen = int(binary.BigEndian.Uint16(resp[payloadLenField:(payloadLenField + payloadLenFieldLen)]))
			// TODO: return error if payloadLen is too large.
		}
	}

	if received == 0 {
		err := errors.New("receiving from serial: timeout, nothing received")
		return nil, err
	}

	// TODO: check if msgStart is present

	if !bytes.Equal(resp[received-len(msgEnd):], msgEnd[:]) {
		err := errors.New("receiving from serial: end of message (0xAA, 0x55) not found")
		return nil, err
	}

	payload := resp[payloadField : payloadField+payloadLen-crcFieldLen]
	ch := crc16.Checksum(payload, crc16.CCITTTable)
	// crc is contained in the last bytes of the payload
	cp := binary.BigEndian.Uint16(resp[payloadField+payloadLen-crcFieldLen : payloadField+payloadLen])
	if ch != cp {
		err := errors.New("receiving from serial: signature of received message is not valid")
		return nil, err
	}

	return payload, nil
}

// encode is internally used to create a valid provisioning packet.
func encode(mType MsgType, msg []byte) []byte {
	// Insert the preamble sequence followed by the message type
	packet := append(msgStart[:], byte(mType))

	// Append the packet length
	bLen := make([]byte, payloadLenFieldLen)
	binary.BigEndian.PutUint16(bLen, (uint16(len(msg) + crcFieldLen)))
	packet = append(packet, bLen...)

	// Append the message payload
	packet = append(packet, msg...)

	// Calculate and append the message signature
	ch := crc16.Checksum(msg, crc16.CCITTTable)
	checksum := make([]byte, crcFieldLen)
	binary.BigEndian.PutUint16(checksum, ch)
	packet = append(packet, checksum...)

	// Append final byte sequence
	packet = append(packet, msgEnd[:]...)
	return packet
}
