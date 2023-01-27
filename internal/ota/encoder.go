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

package ota

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"strconv"

	"github.com/arduino/arduino-cloud-cli/internal/lzss"
)

// Encoder writes a binary to an output stream in the ota format.
type Encoder struct {
	// w is the stream where encoded bytes are written.
	w io.Writer

	// vendorID is the ID of the board vendor.
	magicNumberPart1 string

	// productID is the ID of the board model.
	magicNumberPart2 string
}

// NewEncoder creates a new ota encoder.
func NewEncoder(w io.Writer, magicNumberPart1, magicNumberPart2 string) *Encoder {
	return &Encoder{
		w:                w,
		magicNumberPart1: magicNumberPart1,
		magicNumberPart2: magicNumberPart2,
	}
}

// Encode compresses data using a lzss algorithm, encodes the result
// in ota format and writes it to e's underlying writer.
func (e *Encoder) Encode(data []byte) error {
	// Compute the magic number (VID/PID)
	magicNumber := make([]byte, 4)
	magicNumberPart1, err := strconv.ParseUint(e.magicNumberPart1, 16, 16)
	if err != nil {
		return fmt.Errorf("cannot parse vendorID: %w", err)
	}
	magicNumberPart2, err := strconv.ParseUint(e.magicNumberPart2, 16, 16)
	if err != nil {
		return fmt.Errorf("cannot parse productID: %w", err)
	}

	binary.LittleEndian.PutUint16(magicNumber[0:2], uint16(magicNumberPart2))
	binary.LittleEndian.PutUint16(magicNumber[2:4], uint16(magicNumberPart1))

	// Version field (byte array of size 8)
	version := Version{
		Compression: true,
	}

	compressed := lzss.Encode(data)
	// Prepend magic number and version field to payload
	var outData []byte
	outData = append(outData, magicNumber...)
	outData = append(outData, version.Bytes()...)
	outData = append(outData, compressed...)

	err = e.writeHeader(outData)
	if err != nil {
		return fmt.Errorf("cannot write data header to output stream: %w", err)
	}

	_, err = e.w.Write(outData)
	if err != nil {
		return fmt.Errorf("cannot write encoded data to output stream: %w", err)
	}

	return nil
}

func (e *Encoder) writeHeader(data []byte) error {
	// Write the length of the content
	lengthAsBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthAsBytes, uint32(len(data)))
	_, err := e.w.Write(lengthAsBytes)
	if err != nil {
		return err
	}

	// Write the checksum uint32 value as 4 bytes
	crc := crc32.ChecksumIEEE(data)
	crcAsBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(crcAsBytes, crc)
	_, err = e.w.Write(crcAsBytes)
	if err != nil {
		return err
	}

	return nil
}
