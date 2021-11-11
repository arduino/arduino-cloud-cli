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
	"path/filepath"
	"strings"

	"github.com/arduino/arduino-cloud-cli/internal/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
)

const (
	// default ota should complete in 10 mins
	otaExpirationMins = 10
	// deferred ota can take up to 1 week (equal to 10080 minutes)
	otaDeferredExpirationMins = 10080

	numConcurrentUploads = 10
)

// UploadParams contains the parameters needed to
// perform an OTA upload.
type UploadParams struct {
	DeviceIDs []string
	Tags      map[string]string
	File      string
	Deferred  bool
	FQBN      string
}

// Upload command is used to upload a firmware OTA,
// on a device of Arduino IoT Cloud.
func Upload(params *UploadParams) error {
	if params.DeviceIDs == nil && params.Tags == nil {
		return errors.New("provide either DeviceID or Tags")
	} else if params.DeviceIDs != nil && params.Tags != nil {
		return errors.New("cannot use both DeviceID and Tags. only one of them should be not nil")
	}

	conf, err := config.Retrieve()
	if err != nil {
		return err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return err
	}

	d, err := idsGivenTags(iotClient, params.Tags)
	if err != nil {
		return err
	}
	devs := append(params.DeviceIDs, d...)
	if len(devs) == 0 {
		return errors.New("no device found")
	}

	otaDir, err := ioutil.TempDir("", "")
	if err != nil {
		return fmt.Errorf("%s: %w", "cannot create temporary folder", err)
	}
	otaFile := filepath.Join(otaDir, "temp.ota")
	defer os.RemoveAll(otaDir)

	err = Generate(params.File, otaFile, params.FQBN)
	if err != nil {
		return fmt.Errorf("%s: %w", "cannot generate .ota file", err)
	}

	file, err := os.Open(otaFile)
	if err != nil {
		return fmt.Errorf("%s: %w", "cannot open ota file", err)
	}

	expiration := otaExpirationMins
	if params.Deferred {
		expiration = otaDeferredExpirationMins
	}

	return run(iotClient, devs, file, expiration)
}

func idsGivenTags(iotClient iot.Client, tags map[string]string) ([]string, error) {
	if tags == nil {
		return nil, nil
	}
	devs, err := iotClient.DeviceList(tags)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "cannot retrieve devices from cloud", err)
	}
	devices := make([]string, 0, len(devs))
	for _, d := range devs {
		devices = append(devices, d.Id)
	}
	return devices, nil
}

func run(iotClient iot.Client, ids []string, file *os.File, expiration int) error {
	targets := make(chan string, len(ids))
	type result struct {
		id  string
		err error
	}
	results := make(chan result, len(ids))

	for _, id := range ids {
		targets <- id
	}
	close(targets)

	for i := 0; i < numConcurrentUploads; i++ {
		go func() {
			for id := range targets {
				err := iotClient.DeviceOTA(id, file, expiration)
				results <- result{id: id, err: err}
			}
		}()
	}

	var fails []string
	var details []string
	for range ids {
		r := <-results
		if r.err != nil {
			fails = append(fails, r.id)
			details = append(details, fmt.Sprintf("%s: %s", r.id, r.err.Error()))
		}
	}

	if len(fails) > 0 {
		f := strings.Join(fails, ",")
		f = strings.TrimRight(f, ",")
		d := strings.Join(details, "\n")
		return fmt.Errorf("failed to update these devices: %s\nreasons:\n%s", f, d)
	}
	return nil
}
