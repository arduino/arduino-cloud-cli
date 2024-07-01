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
	"fmt"

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/config"
	storageapi "github.com/arduino/arduino-cloud-cli/internal/storage-api"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

func ApplyCustomTemplates(cred *config.Credentials, templateId string) error {

	apiclient := storageapi.NewClient(cred)

	feedback.Printf("Applying template %s", templateId)

	templateIdUUID, err := uuid.FromString(templateId)
	if err != nil {
		return fmt.Errorf("invalid template id: %s", templateId)
	}
	cstTemplate, err := apiclient.GetCustomTemplate(templateIdUUID)
	if err != nil {
		return err
	}
	if len(cstTemplate.ThingTemplates) > 0 {
		mainThing := cstTemplate.ThingTemplates[0]
		logrus.Debug("Main thing template - id: ", mainThing.Id)
		//TODO check thing ID proceed
	}

	return nil
}
