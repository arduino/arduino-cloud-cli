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
	"context"
	"fmt"
	"os"

	"github.com/arduino/arduino-cloud-cli/config"
)

type EncodeParams struct {
	FQBN string
	File string
}

// Encode command is used to encode a firmware OTA
func Encode(ctx context.Context, params *EncodeParams, cred *config.Credentials) (*string, error) {
	otaFile := fmt.Sprintf("%s.ota", params.File)
	_, err := os.Stat(otaFile)
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
