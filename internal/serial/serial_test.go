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
	"testing"

	"github.com/arduino/arduino-cloud-cli/internal/serial/mocks"
	"github.com/stretchr/testify/mock"
)

func TestSendReceive(t *testing.T) {
	mockPort := &mocks.Port{}
	mockSerial := &Serial{mockPort}

	want := []byte{1, 2, 3}
	resp := encode(Response, want)
	respIdx := 0

	mockRead := func(msg []uint8) int {
		if respIdx >= len(resp) {
			return 0
		}
		copy(msg, resp[respIdx:respIdx+2])
		respIdx += 2
		return 2
	}

	mockPort.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)
	mockPort.On("Read", mock.AnythingOfType("[]uint8")).Return(mockRead, nil)

	res, err := mockSerial.SendReceive(context.TODO(), BeginStorage, []byte{1, 2})
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(res, want) {
		t.Errorf("Expected %v but received %v", want, res)
	}
}

func TestSend(t *testing.T) {
	mockPort := &mocks.Port{}
	mockSerial := &Serial{mockPort}
	mockPort.On("Write", mock.AnythingOfType("[]uint8")).Return(0, nil)

	payload := []byte{1, 2}
	cmd := SetDay
	want := []byte{msgStart[0], msgStart[1], 1, 0, 5, 10, 1, 2, 143, 124, msgEnd[0], msgEnd[1]}

	err := mockSerial.Send(context.TODO(), cmd, payload)
	if err != nil {
		t.Error(err)
	}

	mockPort.AssertCalled(t, "Write", want)
}

func TestEncode(t *testing.T) {
	tests := []struct {
		name string
		msg  []byte
		want []byte
	}{
		{
			name: "begin-storage",
			msg:  []byte{byte(BeginStorage)},
			want: []byte{msgStart[0], msgStart[1], 1, 0, 3, 6, 0x95, 0x4e, msgEnd[0], msgEnd[1]},
		},

		{
			name: "set-year",
			msg:  append([]byte{byte(SetYear)}, []byte("2021")...),
			want: []byte{msgStart[0], msgStart[1], 1, 0, 7, 0x8, 0x32, 0x30, 0x32, 0x31, 0xc3, 0x65, msgEnd[0], msgEnd[1]},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encode(Cmd, tt.msg)
			if !bytes.Equal(tt.want, got) {
				t.Errorf("Expected %v, received %v", tt.want, got)
			}
		})
	}
}
