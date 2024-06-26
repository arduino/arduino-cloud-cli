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

type exportFlags struct {
	templateId string
	path       string
}

func initTemplateExportCommand() *cobra.Command {
	flags := &exportFlags{}
	uploadCommand := &cobra.Command{
		Use:   "export",
		Short: "Export template",
		Long:  "Export template to a file",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runTemplateExportCommand(flags); err != nil {
				feedback.Errorf("Error during template export status: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}

	uploadCommand.Flags().StringVarP(&flags.templateId, "template-id", "t", "", "Template id")
	uploadCommand.Flags().StringVarP(&flags.path, "directory", "d", "", "Output directory")

	uploadCommand.MarkFlagRequired("template-id")

	return uploadCommand
}

func runTemplateExportCommand(flags *exportFlags) error {
	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}
	return template.ExportCustomTemplate(cred, flags.templateId, flags.path)
}
