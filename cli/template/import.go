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

type importFlags struct {
	templateFile string
}

func initTemplateImportCommand() *cobra.Command {
	flags := &importFlags{}
	downloadCommand := &cobra.Command{
		Use:   "import",
		Short: "Import template",
		Long:  "Import a template from a file",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runTemplateImportCommand(flags); err != nil {
				feedback.Errorf("Error during template import: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}

	downloadCommand.Flags().StringVarP(&flags.templateFile, "file", "f", "", "Template file to import")

	downloadCommand.MarkFlagRequired("file")

	return downloadCommand
}

func runTemplateImportCommand(flags *importFlags) error {
	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}
	return template.ImportCustomTemplate(cred, flags.templateFile)
}
