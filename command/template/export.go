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

package template

import (
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/config"
	storageapi "github.com/arduino/arduino-cloud-cli/internal/storage-api"
)

func ExportCustomTemplate(cred *config.Credentials, templateId, path string) error {

	apiclient := storageapi.NewClient(cred)

	feedback.Printf("Exporting template %s", templateId)

	filecreaed, err := apiclient.ExportCustomTemplate(templateId, path)
	if err != nil {
		return err
	}

	outf := ""
	if filecreaed != nil {
		outf = *filecreaed
	}
	feedback.Printf("Template %s exported to file: %s", templateId, outf)

	return nil
}
