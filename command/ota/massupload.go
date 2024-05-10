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
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	"github.com/arduino/arduino-cloud-cli/internal/ota"
	otaapi "github.com/arduino/arduino-cloud-cli/internal/ota-api"

	iotclient "github.com/arduino/iot-client-go"
)

const (
	numConcurrentUploads = 10
)

// MassUploadParams contains the parameters needed to
// perform a Mass OTA upload.
type MassUploadParams struct {
	DeviceIDs        []string
	Tags             map[string]string
	File             string
	Deferred         bool
	DoNotApplyHeader bool
	FQBN             string
}

// Result of an ota upload on a device.
type Result struct {
	ID        string
	Err       error
	OtaStatus otaapi.Ota
}

func buildOtaFile(params *MassUploadParams) (string, string, error) {
	var otaFile string
	var otaDir string
	var err error
	if params.DoNotApplyHeader {
		otaFile = params.File
	} else {
		otaDir, err = os.MkdirTemp("", "")
		if err != nil {
			return "", "", fmt.Errorf("%s: %w", "cannot create temporary folder", err)
		}
		otaFile = filepath.Join(otaDir, "temp.ota")

		err = Generate(params.File, otaFile, params.FQBN)
		if err != nil {
			return "", "", fmt.Errorf("%s: %w", "cannot generate .ota file", err)
		}
	}
	return otaFile, otaDir, nil
}

// MassUpload command is used to mass upload a firmware OTA,
// on devices of Arduino IoT Cloud.
func MassUpload(ctx context.Context, params *MassUploadParams, cred *config.Credentials) ([]Result, error) {
	if params.DeviceIDs == nil && params.Tags == nil {
		return nil, errors.New("provide either DeviceIDs or Tags")
	} else if params.DeviceIDs != nil && params.Tags != nil {
		return nil, errors.New("cannot use both DeviceIDs and Tags. only one of them should be not nil")
	}

	// Generate .ota file
	logrus.Infoln("Uploading binary", params.File)
	_, err := os.Stat(params.File)
	if err != nil {
		return nil, fmt.Errorf("file %s does not exists: %w", params.File, err)
	}

	if !params.DoNotApplyHeader {
		//Verify if file has already an OTA header
		header, _ := ota.DecodeOtaFirmwareHeaderFromFile(params.File)
		if header != nil {
			params.DoNotApplyHeader = true
		}
	}

	// Generate .ota file
	otaFile, otaDir, err := buildOtaFile(params)
	if err != nil {
		return nil, err
	}
	if otaDir != "" {
		defer os.RemoveAll(otaDir)
	}

	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return nil, err
	}
	otapi := otaapi.NewClient(cred)

	// Prepare the list of device-ids to update
	d, err := idsGivenTags(ctx, iotClient, params.Tags)
	if err != nil {
		return nil, err
	}
	d = append(params.DeviceIDs, d...)
	valid, invalid, err := validateDevices(ctx, iotClient, d, params.FQBN)
	if err != nil {
		return nil, fmt.Errorf("failed to validate devices: %w", err)
	}
	if len(valid) == 0 {
		return invalid, nil
	}

	expiration := otaExpirationMins
	if params.Deferred {
		expiration = otaDeferredExpirationMins
	}

	res := run(ctx, iotClient, otapi, valid, otaFile, expiration)
	res = append(res, invalid...)
	return res, nil
}

type deviceLister interface {
	DeviceList(ctx context.Context, tags map[string]string) ([]iotclient.ArduinoDevicev2, error)
}

func idsGivenTags(ctx context.Context, lister deviceLister, tags map[string]string) ([]string, error) {
	if tags == nil {
		return nil, nil
	}
	devs, err := lister.DeviceList(ctx, tags)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "cannot retrieve devices from cloud", err)
	}
	devices := make([]string, 0, len(devs))
	for _, d := range devs {
		devices = append(devices, d.Id)
	}
	return devices, nil
}

func validateDevices(ctx context.Context, lister deviceLister, ids []string, fqbn string) (valid []string, invalid []Result, err error) {
	devs, err := lister.DeviceList(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", "cannot retrieve devices from cloud", err)
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
			inv := Result{ID: id, Err: fmt.Errorf("not found")}
			invalid = append(invalid, inv)
			continue
		}
		// Device FQBN doesn't match the passed one
		if found.Fqbn != fqbn {
			inv := Result{ID: id, Err: fmt.Errorf("has FQBN '%s' instead of '%s'", found.Fqbn, fqbn)}
			invalid = append(invalid, inv)
			continue
		}
		valid = append(valid, id)
	}
	return valid, invalid, nil
}

type otaUploader interface {
	DeviceOTA(ctx context.Context, id string, file *os.File, expireMins int) error
}

type otaStatusGetter interface {
	GetOtaLastStatusByDeviceID(deviceID string) (*otaapi.OtaStatusList, error)
}

func run(ctx context.Context, uploader otaUploader, otapi otaStatusGetter, ids []string, otaFile string, expiration int) []Result {
	type job struct {
		id   string
		file *os.File
	}
	jobs := make(chan job, len(ids))

	resCh := make(chan Result, len(ids))
	results := make([]Result, 0, len(ids))

	for _, id := range ids {
		file, err := os.Open(otaFile)
		if err != nil {
			logrus.Error("cannot open ota file:", otaFile)
			r := Result{ID: id, Err: fmt.Errorf("cannot open ota file")}
			results = append(results, r)
			continue
		}
		defer file.Close()
		jobs <- job{id: id, file: file}
	}
	close(jobs)

	logrus.Infoln("Uploading firmware to devices...")
	for i := 0; i < numConcurrentUploads; i++ {
		go func() {
			for job := range jobs {
				err := uploader.DeviceOTA(ctx, job.id, job.file, expiration)
				otaResult := Result{ID: job.id, Err: err}

				otaID, otaapierr := otapi.GetOtaLastStatusByDeviceID(job.id)
				if otaapierr == nil && otaID != nil && len(otaID.Ota) > 0 {
					otaResult.OtaStatus = otaID.Ota[0]
				}

				resCh <- otaResult
			}
		}()
	}

	for range ids {
		r := <-resCh
		results = append(results, r)
	}
	return results
}
