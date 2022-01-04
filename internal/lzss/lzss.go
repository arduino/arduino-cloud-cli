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

package lzss

import (
	"bytes"
	"io"
)

const (
	ei     = 11              /* typically 10..13 */
	ej     = 4               /* typically 4..5 */
	p      = 1               /* If match length <= P then output one character */
	bufsz  = (1 << ei)       /* buffer size */
	looksz = ((1 << ej) + 1) /* lookahead buffer size */
)

var (
	codecount  = 0
	bit_buffer = 0
	bit_mask   = 128
	EI         = 11              /* typically 10..13 */
	EJ         = 4               /* typically 4..5 */
	P          = 1               /* If match length <= P then output one character */
	N          = (1 << EI)       /* buffer size */
	F          = ((1 << EJ) + 1) /* lookahead buffer size */
)

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

func contains(buf []byte, el []byte) (ok bool, ln int, idx int) {
	for i := 0; i < len(buf)-looksz; i++ {

		// }
		// for i, e := range buf {
		// Skip mismatching elements
		// if el[0] != e {
		if buf[i] != el[0] {
			continue
		}

		// Check bounds
		ahead := min(looksz, len(buf)-i)
		ahead = min(ahead, len(el))

		// Count number of bytes contained
		var j int
		for j = 1; j < ahead; j++ {
			if buf[i+j] != el[j] {
				break
			}
		}
		// store the largest result
		if j > ln {
			ok, ln, idx = true, j, i
		}
	}
	return
}

func putbit1(out io.Writer) {
	bit_buffer |= bit_mask
	bit_mask = bit_mask >> 1
	if bit_mask == 0 {
		out.Write([]byte{byte(bit_buffer)})
		bit_buffer = 0
		bit_mask = 128
	}
}

func putbit0(out io.Writer) {
	bit_mask = bit_mask >> 1
	if bit_mask == 0 {
		out.Write([]byte{byte(bit_buffer)})
		bit_buffer = 0
		bit_mask = 128
	}
}

func flush_bit_buffer(out io.Writer) {
	if bit_mask != 128 {
		out.Write([]byte{byte(bit_buffer)})
	}
}

func output1(out io.Writer, c int) {
	putbit1(out)

	for mask := 256 >> 1; mask != 0; mask = mask >> 1 {
		if c&mask != 0 {
			putbit1(out)
		} else {
			putbit0(out)
		}
	}
}

func output2(out io.Writer, x, y int) {
	putbit0(out)

	for mask := N >> 1; mask != 0; mask = mask >> 1 {
		if x&mask != 0 {
			putbit1(out)
		} else {
			putbit0(out)
		}
	}

	for mask := (1 << EJ) >> 1; mask != 0; mask = mask >> 1 {
		if y&mask != 0 {
			putbit1(out)
		} else {
			putbit0(out)
		}
	}
}

func Encode(data []byte) []byte {
	bit_buffer = 0
	bit_mask = 128
	out := bytes.NewBufferString("")
	in := bytes.NewReader(data)

	var i, j, f1, x, y, r, s, bufferend int
	var c byte

	buffer := make([]byte, N*2)
	for i = 0; i < N-F; i++ {
		buffer[i] = ' '
	}

	for i = N - F; i < N*2; i++ {
		b, err := in.ReadByte()
		if err != nil {
			break
		}
		buffer[i] = b
	}

	bufferend, r, s = i, N-F, 0
	for r < bufferend {
		f1 = min(F, bufferend-r)
		x = 0
		y = 1
		c = buffer[r]
		for i = r - 1; i >= s; i-- {
			if buffer[i] == c {
				for j = 1; j < f1; j++ {
					if buffer[i+j] != buffer[r+j] {
						break
					}
				}
				if j > y {
					x = i
					y = j
				}
			}
		}

		if y <= P {
			output1(out, int(c))
			y = 1
		} else {
			output2(out, x&(N-1), y-2)
		}

		r += y
		s += y
		if r >= N*2-F {
			for i = 0; i < N; i++ {
				buffer[i] = buffer[i+N]
			}
			bufferend -= N
			r -= N
			s -= N

			for bufferend < N*2 {
				b, err := in.ReadByte()
				if err != nil {
					break
				}
				buffer[bufferend] = b
				bufferend++
			}
		}
	}
	flush_bit_buffer(out)

	return out.Bytes()
}
