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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
)

const (
	// default ota should complete in 10 mins
	otaExpirationMins = 10
	// deferred ota can take up to 1 week (equal to 10080 minutes)
	otaDeferredExpirationMins = 10080
)

// UploadParams contains the parameters needed to
// perform an OTA upload.
type UploadParams struct {
	DeviceID string
	File     string
	Deferred bool
}

// Upload command is used to upload a firmware OTA,
// on a device of Arduino IoT Cloud.
func Upload(ctx context.Context, params *UploadParams, cred *config.Credentials) error {
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return err
	}

	dev, err := iotClient.DeviceShow(ctx, params.DeviceID)
	if err != nil {
		return err
	}

	otaDir, err := ioutil.TempDir("", "")
	if err != nil {
		return fmt.Errorf("%s: %w", "cannot create temporary folder", err)
	}
	otaFile := filepath.Join(otaDir, "temp.ota")
	defer os.RemoveAll(otaDir)

	err = Generate(params.File, otaFile, dev.Fqbn)
	if err != nil {
		return fmt.Errorf("%s: %w", "cannot generate .ota file", err)
	}

	file, err := os.Open(otaFile)
	if err != nil {
		return fmt.Errorf("%s: %w", "cannot open ota file", err)
	}
	defer file.Close()

	expiration := otaExpirationMins
	if params.Deferred {
		expiration = otaDeferredExpirationMins
	}

	err = iotClient.DeviceOTA(ctx, params.DeviceID, file, expiration)
	if err != nil {
		return err
	}

	return nil
}
