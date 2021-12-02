package binary

import (
	"bytes"
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

// Download a binary file contained in the binary index
func Download(bin *IndexBin) ([]byte, error) {
	b, err := download(bin.URL)
	if err != nil {
		return nil, fmt.Errorf("cannot download binary at %s: %w", bin.URL, err)
	}

	sz, err := bin.Size.Int64()
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve binary size: %w", err)
	}
	if len(b) != int(sz) {
		return nil, errors.New("download failed: invalid binary size")
	}

	err = verifyChecksum(bin.Checksum, b)
	if err != nil {
		return nil, fmt.Errorf("verifying binary checksum: %w", err)
	}

	return b, nil
}

func download(url string) ([]byte, error) {
	cl := http.Client{
		Timeout: time.Second * 3, // Timeout after 2 seconds
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
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

// verifyChecksum is taken and adapted from:
// https://github.com/arduino/arduino-fwuploader/blob/383d09ce8ecbfbb843272c18437cf4a7a02101e3/indexes/download/download.go#L157
func verifyChecksum(checksum string, file []byte) error {
	if checksum == "" {
		return errors.New("missing checksum")
	}
	split := strings.SplitN(checksum, ":", 2)
	if len(split) != 2 {
		return fmt.Errorf("invalid checksum format: %s", checksum)
	}
	digest, err := hex.DecodeString(split[1])
	if err != nil {
		return fmt.Errorf("invalid hash '%s': %s", split[1], err)
	}

	// names based on: https://docs.oracle.com/javase/8/docs/technotes/guides/security/StandardNames.html#MessageDigest
	var algo hash.Hash
	switch split[0] {
	case "SHA-256":
		algo = crypto.SHA256.New()
	case "SHA-1":
		algo = crypto.SHA1.New()
	case "MD5":
		algo = crypto.MD5.New()
	default:
		return fmt.Errorf("unsupported hash algorithm: %s", split[0])
	}

	if _, err := io.Copy(algo, bytes.NewReader(file)); err != nil {
		return fmt.Errorf("computing hash: %s", err)
	}
	if !bytes.Equal(algo.Sum(nil), digest) {
		return fmt.Errorf("archive hash differs from hash in index")
	}

	return nil
}
