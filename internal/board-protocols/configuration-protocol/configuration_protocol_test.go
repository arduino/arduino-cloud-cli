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

package configurationprotocol

import (
	"errors"
	"testing"

	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/configuration-protocol/cborcoders"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/frame"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/transport"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/transport/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newMockTransportSerial() *mocks.TransportInterface {
	m := &mocks.TransportInterface{}
	m.On("Type").Return(transport.Serial)

	return m
}

func TestNewNetworkConfigurationProtocolSerial(t *testing.T) {
	mockTr := newMockTransportSerial()
	tr := transport.TransportInterface(mockTr)
	ncp := NewNetworkConfigurationProtocol(&tr)
	assert.NotNil(t, ncp)
	assert.Equal(t, &tr, ncp.transport)
	assert.Empty(t, ncp.msgList)
}

func TestConnectSerial_Success(t *testing.T) {
	mockTr := newMockTransportSerial()
	mockTr.On("Connect", mock.Anything).Return(nil)
	mockTr.On("Send", mock.Anything).Return(nil)
	tr := transport.TransportInterface(mockTr)
	ncp := NewNetworkConfigurationProtocol(&tr)
	connectMsg := frame.CreateFrame([]byte{serialInitByte}, frame.TransmissionControl)
	connectParams := transport.TransportInterfaceParams{
		Port:      "COM1",
		BoundRate: 9600,
	}

	err := ncp.Connect("COM1")
	assert.NoError(t, err)
	mockTr.AssertCalled(t, "Connect", connectParams)
	mockTr.AssertCalled(t, "Send", connectMsg.ToBytes())
}

func TestConnectSerial_Error(t *testing.T) {
	mockTr := newMockTransportSerial()
	mockTr.On("Connect", mock.Anything).Return(errors.New("port busy"))
	tr := transport.TransportInterface(mockTr)
	ncp := NewNetworkConfigurationProtocol(&tr)

	err := ncp.Connect("COM1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connecting to serial port")
}

func TestCloseSerial_Success(t *testing.T) {
	mockTr := newMockTransportSerial()
	mockTr.On("Send", mock.Anything).Return(nil)
	mockTr.On("Close").Return(nil)
	tr := transport.TransportInterface(mockTr)
	ncp := NewNetworkConfigurationProtocol(&tr)

	closeMsg := []byte{0x55, 0xaa, 0x03, 0x00, 0x03, 0x02, 0xD3, 0x6A, 0xaa, 0x55}

	err := ncp.Close()
	assert.NoError(t, err)
	mockTr.AssertCalled(t, "Send", closeMsg)
	mockTr.AssertCalled(t, "Close")
	assert.Empty(t, ncp.msgList)
}

func TestSendData_Success(t *testing.T) {
	mockTr := newMockTransportSerial()
	mockTr.On("Connected").Return(true)
	mockTr.On("Send", mock.Anything).Return(nil)
	tr := transport.TransportInterface(mockTr)
	ncp := NewNetworkConfigurationProtocol(&tr)
	connectMessage := cborcoders.From(cborcoders.ProvisioningCommandsMessage{Command: Commands["Connect"]})

	want := []byte{0x55, 0xaa, 0x02, 0x00, 0x09, 0xda, 0x00, 0x01, 0x20, 0x03, 0x81, 0x01, 0x7e, 0x1b, 0xaa, 0x55}
	err := ncp.SendData(connectMessage)
	assert.NoError(t, err)
	mockTr.AssertCalled(t, "Send", want)
}

func TestReceiveData_TransportNotConnected(t *testing.T) {
	mockTr := newMockTransportSerial()
	mockTr.On("Connected").Return(false)
	mockTr.On("Receive", mock.Anything).Return(nil, nil)
	tr := transport.TransportInterface(mockTr)
	ncp := NewNetworkConfigurationProtocol(&tr)

	res, err := ncp.ReceiveData(1)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transport interface is not connected")
}

func TestReceiveData_ReceiveTimeout(t *testing.T) {
	mockTr := newMockTransportSerial()
	mockTr.On("Connected").Return(true)
	mockTr.On("Receive", mock.Anything).Return(nil, errors.New("recv error"))
	tr := transport.TransportInterface(mockTr)
	ncp := NewNetworkConfigurationProtocol(&tr)

	res, err := ncp.ReceiveData(1)
	assert.Nil(t, res)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "recv error")
}

func TestReceiveData_FrameInvalid(t *testing.T) {
	mockTr := newMockTransportSerial()
	invalidFrame := frame.Frame{}
	invalidFrame.SetHeader([]byte{0x55, 0xaa, 0x02, 0x00, 0x03})
	invalidFrame.SetPayload([]byte{0x04})
	invalidFrame.SetCrc([]byte{0x00, 0x00})
	invalidFrame.SetFooter([]byte{0xaa, 0x55})
	want := []byte{0x55, 0xaa, 0x03, 0x00, 0x03, 0x03, 0xC2, 0xE3, 0xaa, 0x55}
	mockTr.On("Connected").Return(true)
	mockTr.On("Receive", mock.Anything).Return([]frame.Frame{invalidFrame}, nil)
	mockTr.On("Send", mock.Anything).Return(nil)
	tr := transport.TransportInterface(mockTr)
	ncp := NewNetworkConfigurationProtocol(&tr)

	res, err := ncp.ReceiveData(1)
	assert.Nil(t, res)
	assert.NoError(t, err)
	mockTr.AssertCalled(t, "Send", want)
}

func TestReceiveData_NackReceived(t *testing.T) {
	mockTr := newMockTransportSerial()
	nackFrame := frame.CreateFrame([]byte{nackByte}, frame.TransmissionControl)
	want := []byte{0x55, 0xaa, 0x02, 0x00, 0x09, 0xda, 0x00, 0x01, 0x20, 0x03, 0x81, 0x01, 0x7e, 0x1b, 0xaa, 0x55}
	mockTr.On("Connected").Return(true)
	mockTr.On("Receive", mock.Anything).Return([]frame.Frame{nackFrame}, nil)
	mockTr.On("Send", mock.Anything).Return(nil)
	tr := transport.TransportInterface(mockTr)
	ncp := NewNetworkConfigurationProtocol(&tr)

	err := ncp.SendData(cborcoders.From(cborcoders.ProvisioningCommandsMessage{Command: Commands["Connect"]}))
	assert.NoError(t, err)
	mockTr.AssertCalled(t, "Send", want)
	res, err := ncp.ReceiveData(1)
	assert.Nil(t, res)
	assert.NoError(t, err)
	mockTr.AssertCalled(t, "Send", want)
}

func TestReceiveData_SerialEndReceived(t *testing.T) {
	mockTr := newMockTransportSerial()
	endFrame := frame.CreateFrame([]byte{serialEndByte}, frame.TransmissionControl)
	mockTr.On("Connected").Return(true)
	mockTr.On("Close").Return(nil)
	mockTr.On("Receive", mock.Anything).Return([]frame.Frame{endFrame}, nil)
	tr := transport.TransportInterface(mockTr)
	ncp := NewNetworkConfigurationProtocol(&tr)

	res, err := ncp.ReceiveData(1)
	assert.Nil(t, res)
	assert.NoError(t, err)
	mockTr.AssertCalled(t, "Close")
}

func TestReceiveData_MsgAndSerialEndReceived(t *testing.T) {
	mockTr := newMockTransportSerial()
	frameData := frame.CreateFrame([]byte{0xda, 0x00, 0x01, 0x20, 0x00, 0x81, 0x01}, frame.Data)
	endFrame := frame.CreateFrame([]byte{serialEndByte}, frame.TransmissionControl)
	mockTr.On("Connected").Return(true)
	mockTr.On("Close").Return(nil)
	mockTr.On("Receive", mock.Anything).Return([]frame.Frame{frameData, endFrame}, nil)
	tr := transport.TransportInterface(mockTr)
	ncp := NewNetworkConfigurationProtocol(&tr)

	res, err := ncp.ReceiveData(1)
	assert.NoError(t, err)
	mockTr.AssertCalled(t, "Close")
	assert.NotNil(t, res)
	assert.Equal(t, res.Type(), cborcoders.ProvisioningStatusMessageType)
	assert.Equal(t, res.ToProvisioningStatusMessage().Status, int16(1))
}

func TestReceiveData_MsgAndNackReceived(t *testing.T) {
	mockTr := newMockTransportSerial()
	frameData := frame.CreateFrame([]byte{0xda, 0x00, 0x01, 0x20, 0x00, 0x81, 0x01}, frame.Data)
	nackFrame := frame.CreateFrame([]byte{nackByte}, frame.TransmissionControl)
	want := []byte{0x55, 0xaa, 0x02, 0x00, 0x09, 0xda, 0x00, 0x01, 0x20, 0x03, 0x81, 0x01, 0x7e, 0x1b, 0xaa, 0x55}
	mockTr.On("Connected").Return(true)
	mockTr.On("Receive", mock.Anything).Return([]frame.Frame{frameData, nackFrame}, nil)
	mockTr.On("Send", mock.Anything).Return(nil)
	tr := transport.TransportInterface(mockTr)
	ncp := NewNetworkConfigurationProtocol(&tr)

	err := ncp.SendData(cborcoders.From(cborcoders.ProvisioningCommandsMessage{Command: Commands["Connect"]}))
	assert.NoError(t, err)
	mockTr.AssertCalled(t, "Send", want)
	res, err := ncp.ReceiveData(1)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, res.Type(), cborcoders.ProvisioningStatusMessageType)
	assert.Equal(t, res.ToProvisioningStatusMessage().Status, int16(1))
	mockTr.AssertCalled(t, "Send", want)
}

func TestReceiveData_ReceiveData(t *testing.T) {
	mockTr := newMockTransportSerial()
	frameData := frame.CreateFrame([]byte{0xda, 0x00, 0x01, 0x20, 0x00, 0x81, 0x01}, frame.Data)
	mockTr.On("Connected").Return(true)
	mockTr.On("Receive", mock.Anything).Return([]frame.Frame{frameData}, nil)
	tr := transport.TransportInterface(mockTr)
	ncp := NewNetworkConfigurationProtocol(&tr)

	res, err := ncp.ReceiveData(1)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, res.Type(), cborcoders.ProvisioningStatusMessageType)
	assert.Equal(t, res.ToProvisioningStatusMessage().Status, int16(1))
}
