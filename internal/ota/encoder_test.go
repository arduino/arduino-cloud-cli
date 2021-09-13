package ota

import (
	"bytes"
	"encoding/hex"
	"log"

	"fmt"
	"hash/crc32"
	"testing"

	"gotest.tools/assert"
)

func TestComputeCrc32Checksum(t *testing.T) {

	data, _ := hex.DecodeString("DEADBEEF")
	crc := crc32.ChecksumIEEE(data)

	assert.Equal(t, crc, uint32(2090640218))
}

func TestEncoderWrite(t *testing.T) {

	// Setup test data
	data, _ := hex.DecodeString("DEADBEEF") // uncompressed, or 'ef 6b 77 de f0' (compressed w/ LZSS)

	var w bytes.Buffer
	vendorID := "2341"  // Arduino
	productID := "8054" // MRK Wifi 1010

	otaWriter := NewWriter(&w, vendorID, productID)
	defer otaWriter.Close()

	n, err := otaWriter.Write(data)
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	log.Println("written ota of", n, "bytes length")

	otaWriter.Close()
	actual := w.Bytes()

	// You can get the expected result creating an `.ota` file using Alex's tools:
	// https://github.com/arduino-libraries/ArduinoIoTCloud/tree/master/extras/tools
	expected, _ := hex.DecodeString("11000000a1744bd4548041230000000000000040ef6b77def0")

	res := bytes.Compare(expected, actual)

	if res != 0 {
		fmt.Println("expected:", hex.Dump(expected), len(expected), "bytes")
		fmt.Println("actual:", hex.Dump(actual), len(actual), "bytes")
	}

	assert.Assert(t, res == 0) // 0 means equal
}
