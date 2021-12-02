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

package binary

import (
	"testing"
)

func TestFindProvisionBin(t *testing.T) {
	var (
		fqbnOK1      = "arduino:samd:nano_33_iot"
		fqbnOK2      = "arduino:samd:mkrwifi1010"
		fqbnNotFound = "arduino:mbed_nano:nano33ble"
	)
	index := &Index{
		Boards: []IndexBoard{
			{Fqbn: fqbnOK1, Provision: &IndexBin{URL: "mkr"}},
			{Fqbn: fqbnOK2, Provision: &IndexBin{URL: "nano"}},
		},
	}

	bin := index.FindProvisionBin(fqbnOK2)
	if bin == nil {
		t.Fatal("provision binary not found")
	}

	bin = index.FindProvisionBin(fqbnNotFound)
	if bin != nil {
		t.Fatalf("provision binary should've not be found, but got: %v", bin)
	}
}

func TestLoadIndex(t *testing.T) {
	_, err := LoadIndex()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
