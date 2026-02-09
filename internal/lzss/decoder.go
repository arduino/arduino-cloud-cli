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

// This code is a go port of LZSS encoder-decoder (Haruhiko Okumura; public domain)

package lzss

import (
	"bytes"

	"github.com/icza/bitio"
)

func Decompress(data []byte) []byte {
	input := bitio.NewReader(bytes.NewBuffer(data))
	output := make([]byte, 0)

	buffer := make([]byte, bufsz*2)
	for i := 0; i < bufsz-looksz; i++ {
		buffer[i] = ' '
	}

	r := bufsz - looksz
	var char byte
	var isChar bool
	var err error
	for {
		isChar, err = input.ReadBool()
		if err != nil {
			break
		}

		if isChar {
			char, err = input.ReadByte()
			if err != nil {
				break
			}
			output = append(output, char)
			buffer[r] = char
			r++
			r &= bufsz - 1
		} else {
			var i, j uint64
			i, err = input.ReadBits(idxsz)
			if err != nil {
				break
			}
			j, err = input.ReadBits(lensz)
			if err != nil {
				break
			}

			for k := 0; k <= int(j)+1; k++ {
				char = buffer[(int(i)+k)&(bufsz-1)]
				output = append(output, char)
				buffer[r] = char
				r++
				r &= bufsz - 1
			}
		}
	}

	return output
}
