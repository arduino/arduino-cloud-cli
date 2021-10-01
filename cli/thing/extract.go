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

package thing

import (
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/thing"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var extractFlags struct {
	id      string
	outfile string
	format  string
}

func initExtractCommand() *cobra.Command {
	extractCommand := &cobra.Command{
		Use:   "extract",
		Short: "Extract a template from a thing",
		Long:  "Extract a template from a Arduino IoT Cloud thing and save it in a file",
		Run:   runExtractCommand,
	}
	extractCommand.Flags().StringVarP(&extractFlags.id, "id", "i", "", "Thing ID")
	extractCommand.Flags().StringVarP(&extractFlags.outfile, "outfile", "o", "", "Template file destination path")
	extractCommand.Flags().StringVar(
		&extractFlags.format,
		"format",
		"yaml",
		"Format of template file, can be {json|yaml}. Default is 'yaml'",
	)

	extractCommand.MarkFlagRequired("id")
	return extractCommand
}

func runExtractCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Extracting template from thing %s\n", extractFlags.id)

	params := &thing.ExtractParams{
		ID:     extractFlags.id,
		Format: extractFlags.format,
	}
	if extractFlags.outfile != "" {
		params.Outfile = &extractFlags.outfile
	}

	err := thing.Extract(params)
	if err != nil {
		feedback.Errorf("Error during template extraction: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Info("Template successfully extracted")
}
