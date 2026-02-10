// This file is part of arduino-cloud-cli.
//
// Copyright (C) ARDUINO SRL (http://www.arduino.cc)
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
	"os"
	"path/filepath"

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	"github.com/arduino/arduino-cloud-cli/internal/ota"
	otaapi "github.com/arduino/arduino-cloud-cli/internal/ota-api"
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
	DeviceID         string
	File             string
	Deferred         bool
	DoNotApplyHeader bool
}

// Upload command is used to upload a firmware OTA,
// on a device of Arduino IoT Cloud.
func Upload(ctx context.Context, params *UploadParams, cred *config.Credentials) error {
	_, err := os.Stat(params.File)
	if err != nil {
		return fmt.Errorf("file %s does not exists: %w", params.File, err)
	}

	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return err
	}
	otapi := otaapi.NewClient(cred)

	dev, err := iotClient.DeviceShow(ctx, params.DeviceID)
	if err != nil {
		return err
	}

	if !params.DoNotApplyHeader {
		//Verify if file has already an OTA header
		header, _ := ota.DecodeOtaFirmwareHeaderFromFile(params.File)
		if header != nil {
			params.DoNotApplyHeader = true
		}
	}

	var otaFile string
	if params.DoNotApplyHeader {
		otaFile = params.File
	} else {
		otaDir, err := os.MkdirTemp("", "")
		if err != nil {
			return fmt.Errorf("%s: %w", "cannot create temporary folder", err)
		}
		otaFile = filepath.Join(otaDir, "temp.ota")
		defer os.RemoveAll(otaDir)

		err = Generate(params.File, otaFile, dereferenceString(dev.Fqbn))
		if err != nil {
			return fmt.Errorf("%s: %w", "cannot generate .ota file", err)
		}
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

	var conflictedOta *otaapi.Ota
	err = iotClient.DeviceOTA(ctx, params.DeviceID, file, expiration)
	if err != nil {
		if errors.Is(err, iot.ErrOtaAlreadyInProgress) {
			conflictedOta = &otaapi.Ota{
				DeviceID:    params.DeviceID,
				Status:      "Skipped",
				ErrorReason: "OTA already in progress",
			}
		} else {
			return err
		}
	}
	// Try to get ota-id from API
	otaID, err := otapi.GetOtaLastStatusByDeviceID(params.DeviceID)
	if err != nil {
		return err
	}
	if otaID != nil && len(otaID.Ota) > 0 {
		if conflictedOta != nil {
			toPrint := otaapi.OtaStatusList{
				Ota: []otaapi.Ota{*conflictedOta, otaID.Ota[0]},
			}
			feedback.PrintResult(toPrint)
		} else {
			feedback.PrintResult(otaID.Ota[0])
		}
	}

	return nil
}
