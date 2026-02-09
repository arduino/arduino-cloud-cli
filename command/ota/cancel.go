// This file is part of arduino-cloud-cli.
//
// Copyright (C) ARDUINO SRL (http://www.arduino.cc/)
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

func CancelOta(otaid string, cred *config.Credentials) error {

	if feedback.GetFormat() == feedback.JSONMini {
		return fmt.Errorf("jsonmini format is not supported for this command")
	}

	otapi := otaapi.NewClient(cred)

	if otaid != "" {
		_, err := otapi.CancelOta(otaid)
		if err != nil {
			return err
		}
		// No error, get current status
		res, err := otapi.GetOtaStatusByOtaID(otaid, 1, otaapi.OrderDesc)
		if err != nil {
			return err
		}
		if res != nil {
			feedback.PrintResult(res.Ota)
		}
		return nil
	}

	return nil
}
