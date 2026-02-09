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

package ota

import (
	"testing"

	"github.com/arduino/arduino-cloud-cli/internal/boardpids"
	"github.com/stretchr/testify/assert"
)

func TestDecodeHeader(t *testing.T) {

	header, err := DecodeOtaFirmwareHeaderFromFile("testdata/cloud.ota")
	assert.Nil(t, err)
	assert.Equal(t, boardpids.ArduinoVendorID, header.VID)
	assert.Equal(t, "8057", header.PID)
	assert.Equal(t, "arduino:samd:nano_33_iot", *header.FQBN)
	assert.Equal(t, boardpids.ArduinoFqbnToPID["arduino:samd:nano_33_iot"], header.PID)

	header, err = DecodeOtaFirmwareHeaderFromFile("testdata/blink.ota")
	assert.Nil(t, err)
	assert.Equal(t, boardpids.ArduinoVendorID, header.VID)
	assert.Equal(t, "8057", header.PID)
	assert.Equal(t, "arduino:samd:nano_33_iot", *header.FQBN)
	assert.Equal(t, boardpids.ArduinoFqbnToPID["arduino:samd:nano_33_iot"], header.PID)

}

func TestDecodeWithNoHeader(t *testing.T) {

	header, err := DecodeOtaFirmwareHeaderFromFile("testdata/cloud.bin")
	assert.Nil(t, header)
	assert.NotNil(t, err)

	header, err = DecodeOtaFirmwareHeaderFromFile("testdata/blink.bin")
	assert.Nil(t, header)
	assert.NotNil(t, err)

}

func TestDecodeEsp32Header(t *testing.T) {

	header, err := DecodeOtaFirmwareHeaderFromFile("testdata/esp32.ota")
	assert.Nil(t, err)
	assert.Equal(t, boardpids.Esp32MagicNumberPart1, header.VID)
	assert.Equal(t, boardpids.Esp32MagicNumberPart2, header.PID)
	assert.Nil(t, header.FQBN)
	assert.Equal(t, "ESP32", header.BoardType)

}
