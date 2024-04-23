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
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/internal/ota"
)

type ReadHeaderParams struct {
	File string
}

// Encode command is used to encode a firmware OTA
func ReadHeader(params *ReadHeaderParams) error {
	_, err := os.Stat(params.File)
	if err != nil {
		return fmt.Errorf("file %s does not exists: %w", params.File, err)
	}

	// Verify if file has already an OTA header
	header, err := ota.DecodeOtaFirmwareHeaderFromFile(params.File)
	if err != nil {
		return fmt.Errorf("file %s does not contains a valid OTA header: %v", params.File, err)
	}
	if header == nil {
		return fmt.Errorf("file %s does not contains a valid OTA header", params.File)
	}

	feedback.PrintResult(header)

	return nil
}
