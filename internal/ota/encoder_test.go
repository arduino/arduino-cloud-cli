// This file is part of arduino-cloud-cli.
//
// Copyright (C) ARDUINO SRL (http://www.arduino.cc/)
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
	"bytes"
	"encoding/hex"
	"os"

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

func TestEncode(t *testing.T) {
	// Setup test data
	data, _ := hex.DecodeString("DEADBEEF") // uncompressed, or 'ef 6b 77 de f0' (compressed w/ LZSS)

	var w bytes.Buffer
	vendorID := "2341"  // Arduino
	productID := "8054" // MRK Wifi 1010

	enc := NewEncoder(&w, vendorID, productID)

	err := enc.Encode(data)
	if err != nil {
		t.Error(err)
	}

	actual := w.Bytes()

	// Expected result has been computed with the following tool:
	// https://github.com/arduino-libraries/ArduinoIoTCloud/tree/master/extras/tools .
	expected, _ := hex.DecodeString("11000000a1744bd4548041230000000000000040ef6b77def0")

	res := bytes.Compare(expected, actual)

	if res != 0 {
		fmt.Println("expected:", hex.Dump(expected), len(expected), "bytes")
		fmt.Println("actual:", hex.Dump(actual), len(actual), "bytes")
	}

	assert.Assert(t, res == 0) // 0 means equal
}

// Expected '.ota' files contained in testdata have been computed with the following tool:
// https://github.com/arduino-libraries/ArduinoIoTCloud/tree/master/extras/tools .
func TestEncodeFiles(t *testing.T) {
	tests := []struct {
		name    string
		infile  string
		outfile string
	}{
		{
			name:    "blink",
			infile:  "testdata/blink.bin",
			outfile: "testdata/blink.ota",
		},
		{
			name:    "cloud sketch",
			infile:  "testdata/cloud.bin",
			outfile: "testdata/cloud.ota",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := os.ReadFile(tt.infile)
			if err != nil {
				t.Fatal("couldn't open test file")
			}

			want, err := os.ReadFile(tt.outfile)
			if err != nil {
				t.Fatal("couldn't open test file")
			}

			var got bytes.Buffer
			vendorID := "2341"  // Arduino
			productID := "8057" // Nano 33 IoT
			otaenc := NewEncoder(&got, vendorID, productID)
			err = otaenc.Encode(input)
			if err != nil {
				t.Error(err)
			}

			if !bytes.Equal(want, got.Bytes()) {
				t.Error("encoding failed")
			}
		})
	}
}
