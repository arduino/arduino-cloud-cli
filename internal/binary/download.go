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
	"context"
	"crypto"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Download a binary file contained in the binary index.
func Download(ctx context.Context, bin *IndexBin) ([]byte, error) {
	b, err := download(ctx, bin.URL)
	if err != nil {
		return nil, fmt.Errorf("cannot download binary at %s: %w", bin.URL, err)
	}

	sz, err := bin.Size.Int64()
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve binary size: %w", err)
	}
	if len(b) != int(sz) {
		return nil, fmt.Errorf("download failed: invalid binary size, expected %d bytes but got %d", sz, len(b))
	}

	err = VerifyChecksum(bin.Checksum, bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("verifying binary checksum: %w", err)
	}

	return b, nil
}

func download(ctx context.Context, url string) ([]byte, error) {
	cl := http.Client{
		Timeout: time.Second * 3,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		err = fmt.Errorf("%s: %w", "request url", err)
		return nil, err
	}
	res, err := cl.Do(req)
	if err != nil {
		err = fmt.Errorf("%s: %w", "do request url", err)
		return nil, err
	}

	if res.Body == nil {
		return nil, errors.New("empty file downloaded")
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		err = fmt.Errorf("%s: %w", "read request body", err)
		return nil, err
	}
	return body, nil
}

// VerifyChecksum has been extracted from github.com/arduino/arduino-fwupload :
// https://github.com/arduino/arduino-fwuploader/blob/fc20a808ece9a082e043e13e3cfa69c571721d76/indexes/download/download.go
//
// We're not using arduino-fwuploader directly as a dependency because it
// indirectly depends on github.com/daaku/go.zipexe which panics during an
// 'init' function, causing cloud-cli to panic when compiled with go1.19.
// More on the issue here: https://github.com/golang/go/issues/54227 .
func VerifyChecksum(checksum string, file io.Reader) error {
	if checksum == "" {
		return errors.New("missing checksum")
	}
	split := strings.SplitN(checksum, ":", 2)
	if len(split) != 2 {
		return fmt.Errorf("invalid checksum format: %s", checksum)
	}
	digest, err := hex.DecodeString(split[1])
	if err != nil {
		return fmt.Errorf("invalid hash '%s': %w", split[1], err)
	}

	// names based on: https://docs.oracle.com/javase/8/docs/technotes/guides/security/StandardNames.html#MessageDigest
	var h hash.Hash
	switch split[0] {
	case "SHA-256":
		h = crypto.SHA256.New()
	case "SHA-1":
		h = crypto.SHA1.New()
	case "MD5":
		h = crypto.MD5.New()
	default:
		return fmt.Errorf("unsupported hash algorithm: %s", split[0])
	}

	if _, err := io.Copy(h, file); err != nil {
		return fmt.Errorf("computing hash: %s", err)
	}
	if !bytes.Equal(h.Sum(nil), digest) {
		return fmt.Errorf("archive hash differs from hash in index")
	}

	return nil
}
