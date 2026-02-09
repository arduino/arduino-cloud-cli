// This file is part of arduino-cloud-cli.
//
// Copyright ARDUINO SRL http://www.arduino.cc/)
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
	"testing"

	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/frame"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/transport"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/transport/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSend_Success(t *testing.T) {
	mockTransportInterface := &mocks.TransportInterface{}
	var tr transport.TransportInterface = mockTransportInterface
	provProt := NewProvisioningProtocol(&tr)
	mockTransportInterface.On("Connected").Return(true)
	mockTransportInterface.On("Send", mock.AnythingOfType("[]uint8")).Return(nil)

	payload := []byte{1, 2}
	cmd := SetDay
	want := []byte{0x55, 0xaa, 1, 0, 5, 10, 1, 2, 143, 124, 0xaa, 0x55}

	err := provProt.Send(context.TODO(), cmd, payload)
	assert.NoError(t, err)
	mockTransportInterface.AssertCalled(t, "Send", want)
}

func TestSendReceive_Success(t *testing.T) {

	mockTransportInterface := &mocks.TransportInterface{}
	var tr transport.TransportInterface = mockTransportInterface
	provProt := NewProvisioningProtocol(&tr)

	want := []byte{1, 2, 3}
	rec := frame.CreateFrame(want, frame.Response)
	receivedListFrame := []frame.Frame{
		rec,
	}

	mockTransportInterface.On("Connected").Return(true)
	mockTransportInterface.On("Send", mock.AnythingOfType("[]uint8")).Return(nil)
	mockTransportInterface.On("Receive", mock.Anything).Return(receivedListFrame, nil)

	res, err := provProt.SendReceive(context.TODO(), BeginStorage, []byte{1, 2})
	assert.NoError(t, err)

	assert.NotNil(t, res, "Expected non-nil response")
	assert.Equal(t, res, want, "Expected %v but received %v", want, res)

}
