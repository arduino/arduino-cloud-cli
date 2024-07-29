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

func ImportCustomTemplate(cred *config.Credentials, filePath string) error {

	apiclient := storageapi.NewClient(cred)

	feedback.Printf("Importing template %s", filePath)

	templateImported, err := apiclient.ImportCustomTemplate(filePath)
	if err != nil {
		return err
	}

	feedback.Printf("Template '%s' (%s) imported successfully", templateImported.Name, templateImported.TemplateId)

	return nil
}
