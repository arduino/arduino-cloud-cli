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
	"os"
	"strconv"
	"strings"
)

var (
	ErrCRC32Mismatch = fmt.Errorf("CRC32 mismatch")
)

type OtaFirmwareHeader struct {
	Length         uint32
	CRC32          uint32
	MagicNumber    uint32
	BoardType      string
	FQBN           *string
	VID            string
	PID            string
	IsArduinoBoard bool
}

func readBytes(file *os.File, length int, offset int64) ([]byte, error) {
	bytes := make([]byte, length)
	_, err := file.ReadAt(bytes, offset)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func crc32Buffered(file *os.File) (uint32, error) {
	h := crc32.NewIEEE()
	// Discard first 8 bytes
	file.Seek(8, 0)
	// Read file in chunks and compute CRC32
	buf := make([]byte, 4096)
	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return 0, err
		}
		if n == 0 {
			break
		}
		h.Write(buf[:n])
	}
	return h.Sum32(), nil
}

func extractXID(buff []byte) string {
	xid := strconv.FormatUint(uint64(binary.LittleEndian.Uint16(buff)), 16)
	return strings.ToUpper(xid)
}

// DecodeOtaFirmwareHeader decodes the OTA firmware header from a binary file.
// File is composed by an header and a payload (optionally lzss compressed).
// Method is also checking CRC32 of the file, verifying that file is not corrupted.
func DecodeOtaFirmwareHeader(binaryFilePath string) (*OtaFirmwareHeader, error) {
	// Check if file exists
	if _, err := os.Stat(binaryFilePath); err != nil {
		return nil, err
	}
	// Open file
	if otafileptr, err := os.Open(binaryFilePath); err != nil {
		return nil, err
	} else {
		defer otafileptr.Close()

		// Get length (payload + header without lenght and CRC32 bytes)
		buff, err := readBytes(otafileptr, 4, 0)
		if err != nil {
			return nil, err
		}
		lenghtInt := binary.LittleEndian.Uint32(buff)

		// Get CRC32 (uint32)
		buff, err = readBytes(otafileptr, 4, 4)
		if err != nil {
			return nil, err
		}
		readsum := binary.LittleEndian.Uint32(buff)

		// Read full binary file (buffered), starting from 8th byte (magic number)
		computedsum, err := crc32Buffered(otafileptr)
		if err != nil {
			return nil, err
		}
		if computedsum != readsum {
			return nil, ErrCRC32Mismatch
		}

		// Get PID+VID (uint32)
		buff, err = readBytes(otafileptr, 8, 8)
		if err != nil {
			return nil, err
		}
		completeMagicNumber := binary.LittleEndian.Uint32(buff)

		//Extract VID and PID
		pid := extractXID(buff[:2])
		vid := extractXID(buff[2:])

		boardType, fqbn, isArduino := getBoardType(completeMagicNumber, pid)

		return &OtaFirmwareHeader{
			Length:         lenghtInt,
			CRC32:          computedsum,
			BoardType:      boardType,
			MagicNumber:    completeMagicNumber,
			IsArduinoBoard: isArduino,
			PID:            pid,
			VID:            vid,
			FQBN:           fqbn,
		}, nil
	}
}

func getBoardType(magicNumber uint32, pid string) (string, *string, bool) {
	baordType := "UNKNOWN"
	if t, ok := BoardTypes[magicNumber]; ok {
		baordType = t
	}
	isArduino := baordType != "UNKNOWN" && baordType != "ESP32"
	var fqbn *string
	if t, ok := ArduinoPidToFQBN[pid]; ok {
		fqbn = &t
	}

	return baordType, fqbn, isArduino
}
