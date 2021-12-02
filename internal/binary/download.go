package binary

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	fwuploader "github.com/arduino/arduino-fwuploader/indexes/download"
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

	err = fwuploader.VerifyChecksum(bin.Checksum, bytes.NewReader(b))
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
