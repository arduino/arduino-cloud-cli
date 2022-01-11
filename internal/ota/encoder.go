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
	"bufio"
	"encoding/binary"
	"hash/crc32"
	"io"
	"strconv"

	"github.com/arduino/arduino-cloud-cli/internal/lzss"
	"github.com/juju/errors"
)

// A writer is a buffered, flushable writer.
type writer interface {
	io.Writer
	Flush() error
}

// encoder encodes a binary into an .ota file.
type encoder struct {
	// w is the writer that compressed bytes are written to.
	w writer

	// vendorID is the ID of the board vendor
	vendorID string

	// is the ID of the board vendor is the ID of the board model
	productID string
}

// NewWriter creates a new `WriteCloser` for the the given VID/PID.
func NewWriter(w io.Writer, vendorID, productID string) io.WriteCloser {
	bw, ok := w.(writer)
	if !ok {
		bw = bufio.NewWriter(w)
	}
	return &encoder{
		w:         bw,
		vendorID:  vendorID,
		productID: productID,
	}
}

// Write writes a compressed representation of p to e's underlying writer.
func (e *encoder) Write(binaryData []byte) (int, error) {
	//log.Println("original binaryData is", len(binaryData), "bytes length")

	// Magic number (VID/PID)
	magicNumber := make([]byte, 4)
	vid, err := strconv.ParseUint(e.vendorID, 16, 16)
	if err != nil {
		return 0, errors.Annotate(err, "OTA encoder: failed to parse vendorID")
	}
	pid, err := strconv.ParseUint(e.productID, 16, 16)
	if err != nil {
		return 0, errors.Annotate(err, "OTA encoder: failed to parse productID")
	}

	binary.LittleEndian.PutUint16(magicNumber[0:2], uint16(pid))
	binary.LittleEndian.PutUint16(magicNumber[2:4], uint16(vid))

	// Version field (byte array of size 8)
	version := Version{
		Compression: true,
	}

	// Compress the compiled binary
	compressed := lzss.Encode(binaryData)

	// Prepend magic number and version field to payload
	var binDataComplete []byte
	binDataComplete = append(binDataComplete, magicNumber...)
	binDataComplete = append(binDataComplete, version.AsBytes()...)
	binDataComplete = append(binDataComplete, compressed...)
	//log.Println("binDataComplete is", len(binDataComplete), "bytes length")

	headerSize, err := e.writeHeader(binDataComplete)
	if err != nil {
		return headerSize, err
	}

	payloadSize, err := e.writePayload(binDataComplete)
	if err != nil {
		return payloadSize, err
	}

	return headerSize + payloadSize, nil
}

// Close closes the encoder, flushing any pending output. It does not close or
// flush e's underlying writer.
func (e *encoder) Close() error {
	return e.w.Flush()
}

func (e *encoder) writeHeader(binDataComplete []byte) (int, error) {

	// Write the length of the content
	lengthAsBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(lengthAsBytes, uint32(len(binDataComplete)))

	n, err := e.w.Write(lengthAsBytes)
	if err != nil {
		return n, err
	}

	// Calculate the checksum for binDataComplete
	crc := crc32.ChecksumIEEE(binDataComplete)

	// encode the checksum uint32 value as 4 bytes
	crcAsBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(crcAsBytes, crc)

	n, err = e.w.Write(crcAsBytes)
	if err != nil {
		return n, err
	}

	return len(lengthAsBytes) + len(crcAsBytes), nil
}

func (e *encoder) writePayload(data []byte) (int, error) {
	return e.w.Write(data)
}
