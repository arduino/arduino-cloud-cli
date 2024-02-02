package ota

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeHeader(t *testing.T) {

	header, err := DecodeOtaFirmwareHeader("testdata/cloud.ota")
	assert.Nil(t, err)
	assert.Equal(t, ArduinoVendorID, header.VID)
	assert.Equal(t, "8057", header.PID)
	assert.Equal(t, "arduino:samd:nano_33_iot", *header.FQBN)
	assert.Equal(t, ArduinoFqbnToPID["arduino:samd:nano_33_iot"], header.PID)

	header, err = DecodeOtaFirmwareHeader("testdata/blink.ota")
	assert.Nil(t, err)
	assert.Equal(t, ArduinoVendorID, header.VID)
	assert.Equal(t, "8057", header.PID)
	assert.Equal(t, "arduino:samd:nano_33_iot", *header.FQBN)
	assert.Equal(t, ArduinoFqbnToPID["arduino:samd:nano_33_iot"], header.PID)

}

func TestDecodeWithNoHeader(t *testing.T) {

	header, err := DecodeOtaFirmwareHeader("testdata/cloud.bin")
	assert.Nil(t, header)
	assert.NotNil(t, err)

}

func TestDecodeEsp32Header(t *testing.T) {

	header, err := DecodeOtaFirmwareHeader("testdata/esp32.ota")
	assert.Nil(t, err)
	assert.Equal(t, Esp32MagicNumberPart1, header.VID)
	assert.Equal(t, Esp32MagicNumberPart2, header.PID)
	assert.Nil(t, header.FQBN)
	assert.Equal(t, "ESP32", header.BoardType)

}
