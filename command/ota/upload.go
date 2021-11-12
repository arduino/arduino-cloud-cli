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

	"github.com/arduino/arduino-cloud-cli/internal/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	iotclient "github.com/arduino/iot-client-go"
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

// UploadResp contains the results of the ota upload
type UploadResp struct {
	Updated []string // Ids of devices updated
	Invalid []string // Ids of device not valid (mismatched fqbn)
	Failed  []string // Ids of device failed
	Errors  []string // Contains detailed errors for each failure
}

// Upload command is used to upload a firmware OTA,
// on a device of Arduino IoT Cloud.
func Upload(params *UploadParams) (*UploadResp, error) {
	if params.DeviceIDs == nil && params.Tags == nil {
		return nil, errors.New("provide either DeviceIDs or Tags")
	} else if params.DeviceIDs != nil && params.Tags != nil {
		return nil, errors.New("cannot use both DeviceIDs and Tags. only one of them should be not nil")
	}

	// Generate .ota file
	otaDir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "cannot create temporary folder", err)
	}
	otaFile := filepath.Join(otaDir, "temp.ota")
	defer os.RemoveAll(otaDir)

	err = Generate(params.File, otaFile, params.FQBN)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "cannot generate .ota file", err)
	}

	conf, err := config.Retrieve()
	if err != nil {
		return nil, err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return nil, err
	}

	d, err := idsGivenTags(iotClient, params.Tags)
	if err != nil {
		return nil, err
	}
	d = append(params.DeviceIDs, d...)
	valid, invalid, details, err := validateDevices(iotClient, d, params.FQBN)
	if err != nil {
		return nil, fmt.Errorf("failed to validate devices: %w", err)
	}
	if len(valid) == 0 {
		return &UploadResp{Invalid: invalid}, nil
	}

	expiration := otaExpirationMins
	if params.Deferred {
		expiration = otaDeferredExpirationMins
	}

	good, fail, ers := run(iotClient, valid, otaFile, expiration)
	if err != nil {
		return nil, err
	}

	// Merge the failure details with the details of invalid devices
	ers = append(details, ers...)

	return &UploadResp{Updated: good, Invalid: invalid, Failed: fail, Errors: ers}, nil
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

func validateDevices(iotClient iot.Client, ids []string, fqbn string) (valid, invalid, details []string, err error) {
	devs, err := iotClient.DeviceList(nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("%s: %w", "cannot retrieve devices from cloud", err)
	}

	for _, id := range ids {
		var found *iotclient.ArduinoDevicev2
		for _, d := range devs {
			if d.Id == id {
				found = &d
				break
			}
		}
		// Device not found on the cloud
		if found == nil {
			invalid = append(invalid, id)
			details = append(details, fmt.Sprintf("%s : not found", id))
			continue
		}
		// Device FQBN doesn't match the passed one
		if found.Fqbn != fqbn {
			invalid = append(invalid, id)
			details = append(details, fmt.Sprintf("%s : has FQBN `%s` instead of `%s`", found.Id, found.Fqbn, fqbn))
			continue
		}
		valid = append(valid, id)
	}
	return valid, invalid, details, nil
}

func run(iotClient iot.Client, ids []string, otaFile string, expiration int) (updated, failed, errors []string) {
	type job struct {
		id   string
		file *os.File
	}
	jobs := make(chan job, len(ids))

	type result struct {
		id  string
		err error
	}
	results := make(chan result, len(ids))

	for _, id := range ids {
		file, err := os.Open(otaFile)
		if err != nil {
			failed = append(failed, id)
			errors = append(errors, fmt.Sprintf("%s: cannot open ota file", id))
		}
		jobs <- job{id: id, file: file}
	}
	close(jobs)

	for i := 0; i < numConcurrentUploads; i++ {
		go func() {
			for job := range jobs {
				err := iotClient.DeviceOTA(job.id, job.file, expiration)
				results <- result{id: job.id, err: err}
			}
		}()
	}

	for range ids {
		r := <-results
		if r.err != nil {
			failed = append(failed, r.id)
			errors = append(errors, fmt.Sprintf("%s: %s", r.id, r.err.Error()))
		} else {
			updated = append(updated, r.id)
		}
	}
	return
}
