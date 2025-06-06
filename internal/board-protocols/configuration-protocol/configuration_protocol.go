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

package configurationprotocol

import (
	"fmt"

	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/configuration-protocol/cborcoders"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/frame"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/transport"
	"github.com/sirupsen/logrus"
)

var StatusBoard = map[int16]string{
	1:    "Connecting",
	2:    "Connected",
	4:    "Resetted",
	100:  "Scanning for WiFi networks",
	-1:   "Failed to connect",
	-3:   "Disconnected",
	-4:   "Parameters not provided",
	-5:   "Invalid parameters",
	-6:   "Cannot execute anew request while another is pending",
	-7:   "Invalid request",
	-8:   "Internet not available",
	-101: "HW Error connectivity module",
	-102: "HW Connectivity Module stopped",
	-150: "Error initializing secure element",
	-151: "Error configuring secure element",
	-152: "Error locking secure element",
	-160: "Error generating UHWID",
	-200: "Error storage begin module",
	-201: "Fail to partition the storage",
	-255: "Generic error",
}

var Commands = map[string]uint8{
	"Connect":                1,
	"GetID":                  2,
	"GetBLEMac":              3,
	"Reset":                  4,
	"ScanWiFi":               100,
	"GetWiFiFWVersion":       101,
	"GetSketchVersion":       200,
	"GetNetConfigLibVersion": 201,
}

const (
	serialInitByte = 0x01
	serialEndByte  = 0x02
	nackByte       = 0x03
)

type NetworkConfigurationProtocol struct {
	transport  *transport.TransportInterface
	msgList    []cborcoders.Cmd
	lastPacket frame.Frame
}

func NewNetworkConfigurationProtocol(transport *transport.TransportInterface) *NetworkConfigurationProtocol {
	return &NetworkConfigurationProtocol{
		transport: transport,
		msgList:   make([]cborcoders.Cmd, 0),
	}
}

func (ncp *NetworkConfigurationProtocol) Connect(address string) error {
	err := (*ncp.transport).Connect(transport.TransportInterfaceParams{
		Port:      address,
		BoundRate: 9600,
	})

	if err != nil {
		err = fmt.Errorf("%s: %w", "connecting to serial port", err)
		return err
	}

	if (*ncp.transport).Type() == transport.Serial {
		p := frame.CreateFrame([]byte{serialInitByte}, frame.TransmissionControl)

		err = (*ncp.transport).Send(p.ToBytes())
	}
	return err

}

func (ncp *NetworkConfigurationProtocol) Close() error {
	if ncp.transport == nil || *ncp.transport == nil {
		return fmt.Errorf("NetworkConfigurationProtocol: transport interface is not initialized")
	}

	if (*ncp.transport).Type() == transport.Serial {
		p := frame.CreateFrame([]byte{serialEndByte}, frame.TransmissionControl)

		err := (*ncp.transport).Send(p.ToBytes())
		if err != nil {
			return fmt.Errorf("error sending end of transmission: %w", err)
		}
	}

	err := (*ncp.transport).Close()
	if err != nil {
		return fmt.Errorf("error closing transport: %w", err)
	}

	ncp.msgList = make([]cborcoders.Cmd, 0)
	ncp.lastPacket = frame.Frame{}

	return nil
}

func (ncp *NetworkConfigurationProtocol) SendData(msg cborcoders.Cmd) error {
	databuf, err := msg.Encode()
	if err != nil {
		return err
	}

	if !(*ncp.transport).Connected() {
		return fmt.Errorf("ProvisioningProtocol: transport interface is not connected")
	}

	packet := frame.CreateFrame(databuf, frame.Data)

	ncp.lastPacket = packet

	return (*ncp.transport).Send(packet.ToBytes())
}

func (ncp *NetworkConfigurationProtocol) ReceiveData(timeoutSeconds int) (*cborcoders.Cmd, error) {
	if (ncp.msgList != nil) && (len(ncp.msgList) > 0) {

		return ncp.popMsg(), nil
	}

	if ncp.transport == nil || *ncp.transport == nil {
		return nil, fmt.Errorf("NetworkConfigurationProtocol: transport interface is not initialized")
	}

	if !(*ncp.transport).Connected() {
		return nil, fmt.Errorf("NetworkConfigurationProtocol: transport interface is not connected")
	}

	frames, err := (*ncp.transport).Receive(timeoutSeconds)
	if err != nil {
		return nil, err
	}

	for _, f := range frames {
		if !f.Validate() {
			ncp.SendNack()
			return nil, nil
		}

		msgType := f.GetType()
		if msgType == frame.TransmissionControl {
			if f.GetPayload()[0] == nackByte {
				// Resend packet
				logrus.Debug("NetworkConfigurationProtocol: Received NACK, resending last packet")
				(*ncp.transport).Send(ncp.lastPacket.ToBytes())
				break
			} else if f.GetPayload()[0] == serialEndByte {
				logrus.Debug("NetworkConfigurationProtocol: Received end of transmission signal")
				(*ncp.transport).Close()
				break
			}
		}
		payload := f.GetPayload()

		res, err := cborcoders.Decode(payload)
		if err != nil {
			logrus.Warnf("NetworkConfigurationProtocol: error decoding payload: %s", err.Error())
			continue
		}

		ncp.msgList = append(ncp.msgList, res)
	}

	return ncp.popMsg(), nil
}

func (ncp *NetworkConfigurationProtocol) SendNack() error {
	packet := frame.CreateFrame([]byte{nackByte}, frame.TransmissionControl)
	return (*ncp.transport).Send(packet.ToBytes())
}

func (ncp *NetworkConfigurationProtocol) popMsg() *cborcoders.Cmd {
	if len(ncp.msgList) == 0 {
		return nil
	}
	msg := ncp.msgList[0]
	ncp.msgList = ncp.msgList[1:]
	return &msg
}
