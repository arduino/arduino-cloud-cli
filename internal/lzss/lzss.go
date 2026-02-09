// This code is a go port of LZSS encoder-decoder (Haruhiko Okumura; public domain)
//
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

package lzss

import (
	"bytes"
)

const (
	idxsz = 11 // Size of buffer indexes in bits, typically 10..13 bits.
	lensz = 4  // Size of lookahead indexes in bits, typically 4..5 bits.

	charsz   = 8   // Size of encoded chars in bits.
	bytemask = 128 // Mask with a bit in 8th position. Used to iterate through bits of a char.

	threshold = 1 // If match length > threshold then output a token (idx, len), otherwise output one char.

	bufsz     = 1 << idxsz       // Buffer size.
	looksz    = (1 << lensz) + 1 // Lookahead buffer size.
	historysz = bufsz - looksz   // History buffer size.

	charStartBit  = true  // Indicates next bits encode a char.
	tokenStartBit = false // Indicates next bits encode a token.
)

// Encode takes a slice of bytes, compresses it using the lzss compression algorithm
// and returns the result in a new bytes buffer.
func Encode(data []byte) []byte {
	// buffer is made up of two parts: the first is for already processed data (history); the second is for new data
	buffer := make([]byte, bufsz*2)
	// Initialize the old-data part (history) of the buffer
	for i := 0; i < historysz; i++ {
		buffer[i] = ' '
	}
	out := newResult()
	in := newFiller(data)

	// Fill the new-data part of the buffer
	n := in.fill(buffer[historysz:])
	bufferend := historysz + n
	for current := historysz; current < bufferend; {
		idx, len := findLargestMatch(buffer, current, bufferend)
		if len <= threshold {
			out.addChar(buffer[current])
			len = 1
		} else {
			out.addToken(idx, len)
		}

		current += len
		if current >= bufsz*2-looksz {
			// Shift processed bytes to the old-data portion of the buffer
			copy(buffer[:bufsz], buffer[bufsz:])
			current -= bufsz
			// Refill the new-data portion of the buffer
			bufferend -= bufsz
			bufferend += in.fill(buffer[bufferend:])
		}
	}

	out.flush()
	return out.bytes()
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// findLargestMatch looks for the largest sequence of characters (from current to current+ahead)
// contained in the history of the buffer.
// It returns the index of the found match, if any, and its length.
// The index is relative to the current position. If idx 0 is returned than no match has been found.
func findLargestMatch(buf []byte, current, size int) (idx, len int) {
	idx = 0
	len = 1
	ahead := min(looksz, size-current)
	history := current - historysz
	c := buf[current]
	for i := current - 1; i >= history; i-- {
		if buf[i] == c {
			var j int
			for j = 1; j < ahead; j++ {
				if buf[i+j] != buf[current+j] {
					break
				}
			}
			if j > len {
				idx = i
				len = j
			}
		}
	}
	return
}

// filler abstracts the process of consuming an input buffer
// using its bytes to fill another buffer.
// It's been used to facilitate the handling of the input buffer in the Encode function.
type filler struct {
	src []byte
	idx int
}

func newFiller(src []byte) *filler {
	return &filler{
		src: src,
	}
}

// fill tries to fill all the dst buffer with bytes read from src.
// It returns the number of bytes moved from src to dst.
// The src buffer offset is then incremented so that all the content of src
// can be consumed in small chunks.
func (f *filler) fill(dst []byte) int {
	n := copy(dst, f.src[f.idx:])
	f.idx += n
	return n
}

// result is responsible for storing the actual result of the encoding.
// It knows how to store characters and tokens in the resulting buffer.
// It must be flushed at the end of the encoding in order to store the
// remaining bits of bitBuffer.
type result struct {
	bitBuffer int
	bitMask   int
	out       *bytes.Buffer
}

func newResult() *result {
	return &result{
		bitBuffer: 0,
		bitMask:   bytemask,
		out:       &bytes.Buffer{},
	}
}

// addChar stores a char in the out buffer.
func (r *result) addChar(c byte) {
	i := int(c)
	r.putbit(charStartBit)
	for mask := (1 << charsz) >> 1; mask != 0; mask = mask >> 1 {
		b := (i & mask) != 0
		r.putbit(b)
	}
}

// addToken stores a token in the out buffer.
func (r *result) addToken(idx, len int) {
	// Adjust idx and len to fit idxsz and lensz bits respectively
	idx &= bufsz - 1
	len -= 2

	r.putbit(tokenStartBit)
	for mask := (1 << idxsz) >> 1; mask != 0; mask = mask >> 1 {
		b := idx&mask != 0
		r.putbit(b)
	}

	for mask := (1 << lensz) >> 1; mask != 0; mask = mask >> 1 {
		b := len&mask != 0
		r.putbit(b)
	}
}

func (r *result) flush() {
	if r.bitMask != bytemask {
		r.out.WriteByte(byte(r.bitBuffer))
	}
}

// putbit puts the passed bit (true -> 1; false -> 0) in the bitBuffer.
// When bitBuffer contains an entire byte it's written to the out buffer.
func (r *result) putbit(b bool) {
	if b {
		r.bitBuffer |= r.bitMask
	}
	r.bitMask = r.bitMask >> 1
	if r.bitMask == 0 {
		r.out.WriteByte(byte(r.bitBuffer))
		r.bitBuffer = 0
		r.bitMask = bytemask
	}
}

func (r *result) bytes() []byte {
	return r.out.Bytes()
}
