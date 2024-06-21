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
	"errors"
	"fmt"
	"os"
	"strings"

	inota "github.com/arduino/arduino-cloud-cli/internal/ota"
)

// Generate takes a .bin file and generates a .ota file.
func Generate(binFile string, outFile string, fqbn string) error {

	// We are going to put a magic number in the ota .bin file, the fw will check the magic number once the binary is received
	var magicNumberPart1, magicNumberPart2 string

	// The ota update is available for Arduino boards and ESP32 boards

	// Esp32 boards have a wide range of vid and pid, we don't map all of them
	// If the fqbn is the one of an ESP32 board, we force a default magic number that matches the same default expected on the fw side
	if !strings.HasPrefix(fqbn, "arduino:esp32") && strings.HasPrefix(fqbn, "esp32") {
		magicNumberPart1 = inota.Esp32MagicNumberPart1
		magicNumberPart2 = inota.Esp32MagicNumberPart2
	} else {
		//For Arduino Boards we use vendorId and productID to form the magic number
		magicNumberPart1 = inota.ArduinoVendorID
		productID, ok := inota.ArduinoFqbnToPID[fqbn]
		if !ok {
			return errors.New("fqbn not valid")
		}
		magicNumberPart2 = productID
	}

	data, err := os.ReadFile(binFile)
	if err != nil {
		return err
	}

	out, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer out.Close()

	enc := inota.NewEncoder(out, magicNumberPart1, magicNumberPart2)
	err = enc.Encode(data)
	if err != nil {
		return fmt.Errorf("failed to encode binary file: %w", err)
	}

	return nil
}

func dereferenceString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
