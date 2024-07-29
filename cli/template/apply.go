// This file is part of arduino-cloud-cli.
//
// Copyright (C) 2024 ARDUINO SA (http://www.arduino.cc/)
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
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/template"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/spf13/cobra"
)

type applyFlags struct {
	templateId     string
	templatePrefix string
	deviceId       string
	netCredentials string
	applyOta       bool
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

	applyCommand.Flags().StringVarP(&flags.templateId, "template-id", "t", "", "Template ID")
	applyCommand.Flags().StringVarP(&flags.templatePrefix, "prefix", "p", "", "Prefix to apply to the name of created resources")
	applyCommand.Flags().StringVarP(&flags.deviceId, "device-id", "d", "", "Device ID")
	applyCommand.Flags().StringVarP(&flags.netCredentials, "network-credentials", "n", "", "Comma separated network credentials used to configure device with format <key>=<value>. Supported values: SECRET_SSID | SECRET_OPTIONAL_PASS | SECRET_DEVICE_KEY")

	applyCommand.MarkFlagRequired("template-id")
	applyCommand.MarkFlagRequired("prefix")
	applyCommand.MarkFlagRequired("device-id")

	flags.applyOta = false

	return applyCommand
}

func runTemplateApplyCommand(flags *applyFlags) error {
	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	deviceNetCredentials, err := parseCredentials(flags.netCredentials)
	if err != nil {
		return fmt.Errorf("parsing network credentials: %w", err)
	}

	return template.ApplyCustomTemplates(cred, flags.templateId, flags.deviceId, flags.templatePrefix, deviceNetCredentials, flags.applyOta)
}

func parseCredentials(credentials string) (map[string]string, error) {
	credentialsMap := make(map[string]string)
	if credentials == "" {
		return credentialsMap, nil
	}
	credentialsArray := strings.Split(credentials, ",")
	for _, credential := range credentialsArray {
		credentialArray := strings.Split(credential, "=")
		if len(credentialArray) != 2 {
			return nil, fmt.Errorf("invalid network credential: %s", credential)
		}
		credentialsMap[credentialArray[0]] = credentialArray[1]
	}
	return credentialsMap, nil
}
