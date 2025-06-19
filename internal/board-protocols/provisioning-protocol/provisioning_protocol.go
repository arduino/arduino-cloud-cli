// This file is part of arduino-cloud-cli.
//
// Copyright (C) 2025 ARDUINO SA (http://www.arduino.cc/)
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

package provisioningprotocol

import (
	"context"
	"errors"
	"fmt"

	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/frame"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/transport"
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

const (
	timeoutSeconds = 2
)

type ProvisioningProtocol struct {
	transport  *transport.TransportInterface
	packetList []frame.Frame
}

func NewProvisioningProtocol(transport *transport.TransportInterface) *ProvisioningProtocol {
	return &ProvisioningProtocol{
		transport: transport,
	}
}

// Send allows to send a provisioning command to a connected arduino device.
func (p *ProvisioningProtocol) Send(ctx context.Context, cmd Command, payload []byte) error {
	if p.transport == nil || *p.transport == nil {
		return fmt.Errorf("ProvisioningProtocol: transport interface is not initialized")
	}

	if !(*p.transport).Connected() {
		return fmt.Errorf("ProvisioningProtocol: transport interface is not connected")
	}

	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context error: %w", err)
	}
	payload = append([]byte{byte(cmd)}, payload...)
	frame := frame.CreateFrame(payload, frame.Cmd)

	err := (*p.transport).Send(frame.ToBytes())
	return err

}

// SendReceive allows to send a provisioning command to a connected arduino device.
// Then, it waits for a response from the device and, if any, returns it.
// If no response is received after 2 seconds, an error is returned.
func (p *ProvisioningProtocol) SendReceive(ctx context.Context, cmd Command, payload []byte) ([]byte, error) {
	if err := p.Send(ctx, cmd, payload); err != nil {
		return nil, err
	}
	return p.receive(ctx)
}

// receive allows to wait for a response from an arduino device under provisioning.
// Its timeout is set to 2 seconds. It returns an error if the response is not valid
// or if the timeout expires.
func (p *ProvisioningProtocol) receive(ctx context.Context) ([]byte, error) {
	if p.packetList != nil || len(p.packetList) > 0 {
		return p.popMsg().GetPayload(), nil
	}

	if p.transport == nil || *p.transport == nil {
		return nil, errors.New("ProvisioningProtocol: transport interface is not initialized")
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	packets, err := (*p.transport).Receive(timeoutSeconds)

	if err != nil {
		return nil, fmt.Errorf("error receiving packets: %w", err)
	}

	if len(packets) == 0 {
		return nil, nil
	}

	for _, packet := range packets {
		if !packet.Validate() {
			return nil, errors.New("received invalid packet")
		}
	}

	p.packetList = append(p.packetList, packets...)

	return p.popMsg().GetPayload(), nil
}

func (p *ProvisioningProtocol) popMsg() *frame.Frame {
	if len(p.packetList) == 0 {
		return nil
	}
	msg := p.packetList[0]
	p.packetList = p.packetList[1:]
	return &msg
}
