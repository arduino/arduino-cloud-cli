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
	applyCommand.Flags().StringVarP(&flags.netCredentials, "network-credentials", "n", "", "Network credentials")

	applyCommand.MarkFlagRequired("template-id")
	applyCommand.MarkFlagRequired("prefix")
	applyCommand.MarkFlagRequired("device-id")

	return applyCommand
}

func runTemplateApplyCommand(flags *applyFlags) error {
	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	deviceNetCredentials := make(map[string]string)
	if flags.netCredentials != "" {
		configNetArray := strings.Split(strings.Trim(flags.netCredentials, " "), ",")
		for _, netConfig := range configNetArray {
			netConfigArray := strings.Split(netConfig, "=")
			if len(netConfigArray) != 2 {
				return fmt.Errorf("invalid network configuration: %s", netConfig)
			}
			deviceNetCredentials[netConfigArray[0]] = netConfigArray[1]
		}
	}

	return template.ApplyCustomTemplates(cred, flags.templateId, flags.deviceId, flags.templatePrefix, deviceNetCredentials)
}
