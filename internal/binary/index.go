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
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/arduino/arduino-cloud-cli/internal/binary/gpgkey"
	"golang.org/x/crypto/openpgp"
)

const (
	// URL of cloud-team binary index
	BinaryIndexURL = "https://cloud-downloads.arduino.cc/binaries/index.json"
	// URL of binary index signature
	BinaryIndexSigURL = "https://cloud-downloads.arduino.cc/binaries/index.json.sig"
)

// Index contains details about all the binaries
// loaded in 'cloud-downloads'
type Index struct {
	Boards []IndexBoard
}

// IndexBoard describes all the binaries available for a specific board
type IndexBoard struct {
	Fqbn      string    `json:"fqbn"`
	Provision *IndexBin `json:"provision"`
}

// IndexBin contains the details needed to retrieve a binary file from the cloud
type IndexBin struct {
	URL      string      `json:"url"`
	Checksum string      `json:"checksum"`
	Size     json.Number `json:"size"`
}

// LoadIndex downloads and verify the index of binaries contained
// in 'cloud-downloads'
func LoadIndex() (*Index, error) {
	index, err := download(BinaryIndexURL)
	if err != nil {
		return nil, fmt.Errorf("cannot download index: %w", err)
	}

	sig, err := download(BinaryIndexSigURL)
	if err != nil {
		return nil, fmt.Errorf("cannot download index signature: %w", err)
	}

	keyRing, err := openpgp.ReadKeyRing(bytes.NewReader(gpgkey.IndexPublicKey))
	if err != nil {
		return nil, fmt.Errorf("retrieving Arduino public key: %w", err)
	}

	signer, err := openpgp.CheckDetachedSignature(keyRing, bytes.NewReader(index), bytes.NewReader(sig))
	if signer == nil || err != nil {
		return nil, fmt.Errorf("index at %s not valid", BinaryIndexURL)
	}

	i := &Index{}
	err = json.Unmarshal(index, &i.Boards)
	return i, err
}

// FindProvisionBin looks for the provisioning binary corresponding
// to the passed fqbn in the index.
// Returns nil if the binary is not found
func (i *Index) FindProvisionBin(fqbn string) *IndexBin {
	for _, b := range i.Boards {
		if b.Fqbn == fqbn {
			return b.Provision
		}
	}
	return nil
}
