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
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/template"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/spf13/cobra"
)

type applyFlags struct {
	templateId   string
	templateName string
}

func initTemplateApplyCommand() *cobra.Command {
	flags := &applyFlags{}
	applyCommand := &cobra.Command{
		Use:   "apply",
		Short: "Apply custom template",
		Long:  "Given a template, apply it and create all the resources defined in it",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runTemplateApplyCommand(flags); err != nil {
				feedback.Errorf("Error during template apply: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}

	applyCommand.Flags().StringVarP(&flags.templateId, "template-id", "t", "", "Template id")
	applyCommand.Flags().StringVarP(&flags.templateName, "name", "n", "", "Name")

	applyCommand.MarkFlagRequired("template-id")
	applyCommand.MarkFlagRequired("name")

	return applyCommand
}

func runTemplateApplyCommand(flags *applyFlags) error {
	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}
	return template.ApplyCustomTemplates(cred, flags.templateId)
}

/*

curl --location --request PUT 'http://localhost:9000/iot/v1/templates' \
--header 'Accept: application/yaml' \
--header 'Content-Type: application/json' \
--header 'Authorization: ' \
--data '{
    "template_name": "home",
    "custom_template_id": "d864f20e-dcf4-4c8a-b3e7-f1bfffe86f60",
    "things_options": {
        "home-3a06e": {
            "device_id": "08d75172-335e-4cb9-b401-83eb4db213fb",
            "secrets": {
                "SECRET_SSID": "asdas",
                "SECRET_OPTIONAL_PASS": "asdsad"
            }
        }
    }
}'

*/
