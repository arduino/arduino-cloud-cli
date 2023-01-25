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
	"io/ioutil"
	"os"

	inota "github.com/arduino/arduino-cloud-cli/internal/ota"
)

var (
	arduinoVendorID = "2341"
	fqbnToPID       = map[string]string{
		"arduino:samd:nano_33_iot":            "8057",
		"arduino:samd:mkr1000":                "804E",
		"arduino:samd:mkrgsm1400":             "8052",
		"arduino:samd:mkrnb1500":              "8055",
		"arduino:samd:mkrwifi1010":            "8054",
		"arduino:mbed_nano:nanorp2040connect": "005E",
		"arduino:mbed_portenta:envie_m7":      "025B",
		"arduino:mbed_nicla:nicla_vision":     "025F",
		"arduino:mbed_opta:opta":              "0064",
	}
)

// Generate takes a .bin file and generates a .ota file.
func Generate(binFile string, outFile string, fqbn string) error {
	productID, ok := fqbnToPID[fqbn]
	if !ok {
		return errors.New("fqbn not valid")
	}

	data, err := ioutil.ReadFile(binFile)
	if err != nil {
		return err
	}

	out, err := os.Create(outFile)
	if err != nil {
		return err
	}
	defer out.Close()

	enc := inota.NewEncoder(out, arduinoVendorID, productID)
	err = enc.Encode(data)
	if err != nil {
		return fmt.Errorf("failed to encode binary file: %w", err)
	}

	return nil
}
