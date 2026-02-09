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

package lzss

import (
	"bytes"
	"os"
	"testing"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		infile  string
		outfile string
	}{
		{
			name:    "blink",
			infile:  "testdata/blink.lzss",
			outfile: "testdata/blink.bin",
		},
		{
			name:    "cloud sketch",
			infile:  "testdata/cloud.lzss",
			outfile: "testdata/cloud.bin",
		},
		{
			name:    "empty binary",
			infile:  "testdata/empty.lzss",
			outfile: "testdata/empty.bin",
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

			got := Decompress(input)
			if !bytes.Equal(want, got) {
				t.Error("decoding failed", want, got)
			}
		})
	}
}
