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
	"fmt"

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/config"
	otaapi "github.com/arduino/arduino-cloud-cli/internal/ota-api"
)

func PrintOtaStatus(otaid, otaids, device string, cred *config.Credentials, limit int, order string) error {

	if feedback.GetFormat() == feedback.JSONMini {
		return fmt.Errorf("jsonmini format is not supported for this command")
	}

	otapi := otaapi.NewClient(cred)

	if otaids != "" {
		res, err := otapi.GetOtaStatusByOtaIDs(otaids)
		if err == nil && res != nil {
			feedback.PrintResult(res)
		} else if err != nil {
			return err
		}
	} else if otaid != "" {
		res, err := otapi.GetOtaStatusByOtaID(otaid, limit, order)
		if err == nil && res != nil {
			feedback.PrintResult(otaapi.OtaStatusDetail{
				FirmwareSize: res.Ota.FirmwareSize,
				Ota:          res.Ota,
				Details:      res.States,
				MaxRetries:   res.Ota.MaxRetries,
				RetryAttempt: res.Ota.RetryAttempt,
			})
		} else if err != nil {
			return err
		}
	} else if device != "" {
		res, err := otapi.GetOtaStatusByDeviceID(device, limit, order)
		if err == nil && res != nil {
			feedback.PrintResult(res)
		} else if err != nil {
			return err
		}
	}

	return nil
}
