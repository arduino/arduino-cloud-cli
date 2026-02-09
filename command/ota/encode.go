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
	"fmt"
	"os"
	"strings"

	"github.com/arduino/arduino-cloud-cli/internal/ota"
)

type EncodeParams struct {
	FQBN string
	File string
}

// Encode command is used to encode a firmware OTA
func Encode(params *EncodeParams) (*string, error) {
	_, err := os.Stat(params.File)
	if err != nil {
		return nil, fmt.Errorf("file %s does not exists: %w", params.File, err)
	}

	// Verify if file has already an OTA header
	header, _ := ota.DecodeOtaFirmwareHeaderFromFile(params.File)
	if header != nil {
		return nil, fmt.Errorf("file %s contains a valid OTA header. Skip header encoding", params.File)
	}

	var otaFile string
	if strings.HasSuffix(params.File, ".bin") {
		otaFile = strings.Replace(params.File, ".bin", ".ota", 1)
	} else {
		otaFile = fmt.Sprintf("%s.ota", params.File)
	}
	_, err = os.Stat(otaFile)
	if err == nil {
		// file already exists, we need to delete it
		if err = os.Remove(otaFile); err != nil {
			return nil, fmt.Errorf("%s: %w", "cannot remove .ota file", err)
		}
	}

	err = Generate(params.File, otaFile, params.FQBN)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "cannot generate .ota file", err)
	}

	return &otaFile, nil
}
