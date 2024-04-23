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
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/arduino-cloud-cli/internal/lzss"
)

var (
	ErrCRC32Mismatch  = fmt.Errorf("CRC32 mismatch")
	ErrLengthMismatch = fmt.Errorf("file length mismatch")

	boardTypes = map[uint32]string{
		0x45535033: "ESP32",
		0x23418054: "MKR_WIFI_1010",
		0x23418057: "NANO_33_IOT",
		0x2341025B: "PORTENTA_H7_M7",
		0x2341005E: "NANO_RP2040_CONNECT",
		0x2341025F: "NICLA_VISION",
		0x23410064: "OPTA",
		0x23410266: "GIGA",
		0x23410070: "NANO_ESP32",
		0x23411002: "UNOR4WIFI",
	}

	arduinoPidToFQBN = map[string]string{
		"8057": "arduino:samd:nano_33_iot",
		"804E": "arduino:samd:mkr1000",
		"8052": "arduino:samd:mkrgsm1400",
		"8055": "arduino:samd:mkrnb1500",
		"8054": "arduino:samd:mkrwifi1010",
		"005E": "arduino:mbed_nano:nanorp2040connect",
		"025B": "arduino:mbed_portenta:envie_m7",
		"025F": "arduino:mbed_nicla:nicla_vision",
		"0064": "arduino:mbed_opta:opta",
		"0266": "arduino:mbed_giga:giga",
		"0070": "arduino:esp32:nano_nora",
		"1002": "arduino:renesas_uno:unor4wifi",
	}
)

const (
	OffsetLength      = 0
	OffsetCRC32       = 4
	OffsetMagicNumber = 8
	OffsetVersion     = 12
	OffsetPayload     = 20
	HeaderSize        = 20
)

type OtaFileReader interface {
	io.Reader
	io.Closer
}

type OtaMetadata struct {
	Length         uint32
	CRC32          uint32
	MagicNumber    uint32
	BoardType      string
	FQBN           *string
	VID            string
	PID            string
	IsArduinoBoard bool
	Compressed     bool
	PayloadSHA256  string // SHA256 of the payload (decompressed if compressed). This is the SHA256 as seen ny the board.
	OtaSHA256      string // SHA256 of the whole file (header + payload).
}

func (r OtaMetadata) Data() interface{} {
	return r
}

func (r OtaMetadata) String() string {
	t := table.New()

	t.SetHeader("Entry", "Value")

	t.AddRow([]interface{}{"Length", fmt.Sprintf("%d bytes", r.Length)}...)
	t.AddRow([]interface{}{"CRC32", fmt.Sprintf("%d", r.CRC32)}...)
	t.AddRow([]interface{}{"Magic Number", fmt.Sprintf("0x%08X", r.MagicNumber)}...)
	t.AddRow([]interface{}{"Board Type", r.BoardType}...)
	if r.FQBN != nil {
		t.AddRow([]interface{}{"FQBN", *r.FQBN}...)
	}
	t.AddRow([]interface{}{"VID", r.VID}...)
	t.AddRow([]interface{}{"PID", r.PID}...)
	t.AddRow([]interface{}{"Is Arduino Board", strconv.FormatBool(r.IsArduinoBoard)}...)
	t.AddRow([]interface{}{"Compressed", strconv.FormatBool(r.Compressed)}...)
	t.AddRow([]interface{}{"Payload SHA256", r.PayloadSHA256}...)
	t.AddRow([]interface{}{"OTA SHA256", r.OtaSHA256}...)

	return t.Render()
}

// Read header starting from the first byte of the file
func readHeader(file OtaFileReader) ([]byte, error) {
	bytes := make([]byte, HeaderSize)
	_, err := file.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// Function will compute OTA CRC32 and file SHA256 hash, starting from a reader that has already extracted the header (so pointing to payload).
func computeFileHashes(file OtaFileReader, compressed bool, otaHeader []byte) (uint32, string, string, uint32, error) {
	crcSum := crc32.NewIEEE()
	payload := bytes.Buffer{}
	// Length of remaining header + payload excluding the fields LENGTH and CRC32.
	computedLength := HeaderSize - OffsetMagicNumber

	// Discard first 8 bytes (len + crc32) and read remaining header's bytes (12B - magic number + version)
	crcSum.Write(otaHeader[8:HeaderSize])

	// Read file in chunks and compute CRC32. Save payload in a buffer for next processing steps (SHA256)
	buf := make([]byte, 4096)
	for {
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			return 0, "", "", 0, err
		}
		if n == 0 {
			break
		}
		computedLength += n
		crcSum.Write(buf[:n])
		payload.Write(buf[:n])
	}

	payloadSHA, otaSHA := computeBinarySha256(compressed, payload.Bytes(), otaHeader)

	return crcSum.Sum32(), payloadSHA, otaSHA, uint32(computedLength), nil
}

func computeBinarySha256(compressed bool, payload []byte, otaHeader []byte) (string, string) {
	var computedShaBytes [32]byte
	if compressed {
		decompressed := lzss.Decompress(payload)
		computedShaBytes = sha256.Sum256(decompressed)
	} else {
		computedShaBytes = sha256.Sum256(payload)
	}

	// Whole file SHA256 (header + payload)
	otaSHA := sha256.New()
	otaSHA.Write(otaHeader)
	otaSHA.Write(payload)

	return hex.EncodeToString(computedShaBytes[:]), hex.EncodeToString(otaSHA.Sum(nil))
}

func extractXID(buff []byte) string {
	xid := strconv.FormatUint(uint64(binary.LittleEndian.Uint16(buff)), 16)
	return strings.ToUpper(xid)
}

// DecodeOtaFirmwareHeader decodes the OTA firmware header from a binary file.
// File is composed by an header and a payload (optionally lzss compressed).
// Method is also checking CRC32 of the file, verifying that file is not corrupted.
// OTA header layout: LENGTH (4 B) | CRC (4 B) | MAGIC NUMBER = VID + PID (4 B) | VERSION (8 B) | PAYLOAD (LENGTH - 12 B)
// See https://arduino.atlassian.net/wiki/spaces/RFC/pages/1616871540/OTA+header+structure
func DecodeOtaFirmwareHeaderFromFile(binaryFilePath string) (*OtaMetadata, error) {
	// Check if file exists
	if _, err := os.Stat(binaryFilePath); err != nil {
		return nil, err
	}
	if otafileptr, err := os.Open(binaryFilePath); err != nil {
		return nil, err
	} else {
		defer otafileptr.Close()
		return DecodeOtaFirmwareHeader(otafileptr)
	}
}

// DecodeOtaFirmwareHeader decodes the OTA firmware header from a binary file.
// File is composed by an header and a payload (optionally lzss compressed).
// Method is also checking CRC32 of the file, verifying that file is not corrupted.
// OTA header layout: LENGTH (4 B) | CRC (4 B) | MAGIC NUMBER = VID + PID (4 B) | VERSION (8 B) | PAYLOAD (LENGTH - 12 B)
// See https://arduino.atlassian.net/wiki/spaces/RFC/pages/1616871540/OTA+header+structure
func DecodeOtaFirmwareHeader(otafileptr OtaFileReader) (*OtaMetadata, error) {
	header, err := readHeader(otafileptr) // Read all header.
	if err != nil {
		return nil, err
	}

	// Get length (payload + header without length and CRC32 bytes)
	lengthInt := binary.LittleEndian.Uint32(header[OffsetLength:OffsetCRC32])

	// Get CRC32 (uint32)
	readsum := binary.LittleEndian.Uint32(header[OffsetCRC32:OffsetMagicNumber])

	// Get PID+VID (uint32)
	completeMagicNumber := binary.LittleEndian.Uint32(header[OffsetMagicNumber:OffsetVersion])

	//Extract PID and VID. VID is in the last 2 bytes of the magic number, PID in the first 2 bytes.
	pid := extractXID(header[OffsetMagicNumber : OffsetMagicNumber+2])
	vid := extractXID(header[OffsetMagicNumber+2 : OffsetVersion])

	boardType, fqbn, isArduino := getBoardType(completeMagicNumber, pid)

	// Get Version (8B)
	version := decodeVersion(header[OffsetVersion:OffsetPayload])

	// Read full binary file (buffered), starting from 8th byte (magic number)
	computedsum, fileSha, otaSha, computedLength, err := computeFileHashes(otafileptr, version.Compression, header)
	if err != nil {
		return nil, err
	}

	// File sanity check. Validate CRC32 and length declared in header with computed values.
	if computedsum != readsum {
		return nil, ErrCRC32Mismatch
	}
	if computedLength != lengthInt {
		return nil, ErrLengthMismatch
	}

	return &OtaMetadata{
		Length:         lengthInt,
		CRC32:          computedsum,
		BoardType:      boardType,
		MagicNumber:    completeMagicNumber,
		IsArduinoBoard: isArduino,
		PID:            pid,
		VID:            vid,
		FQBN:           fqbn,
		Compressed:     version.Compression,
		PayloadSHA256:  fileSha,
		OtaSHA256:      otaSha,
	}, nil
}

func getBoardType(magicNumber uint32, pid string) (string, *string, bool) {
	baordType := "UNKNOWN"
	if t, ok := boardTypes[magicNumber]; ok {
		baordType = t
	}
	isArduino := baordType != "UNKNOWN" && baordType != "ESP32"
	var fqbn *string
	if t, ok := arduinoPidToFQBN[pid]; ok {
		fqbn = &t
	}

	return baordType, fqbn, isArduino
}
