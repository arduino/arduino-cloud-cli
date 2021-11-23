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

package config

import (
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var initFlags struct {
	destDir   string
	overwrite bool
	format    string
}

func initInitCommand() *cobra.Command {
	initCommand := &cobra.Command{
		Use:   "init",
		Short: "Initialize a configuration file with default values",
		Long:  "Initialize an Arduino IoT Cloud CLI configuration file with default values",
		Run:   runInitCommand,
	}

	initCommand.Flags().StringVar(&initFlags.destDir, "dest-dir", "", "Sets where to save the configuration file")
	initCommand.Flags().BoolVar(&initFlags.overwrite, "overwrite", false, "Overwrite existing config file")
	initCommand.Flags().StringVar(&initFlags.format, "config-format", "yaml", "Format of the configuration file, can be {yaml|json}")

	return initCommand
}

func runInitCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Initializing config file")

	params := &config.InitParams{
		DestDir:   initFlags.destDir,
		Overwrite: initFlags.overwrite,
		Format:    initFlags.format,
	}

	file, err := config.Init(params)
	if err != nil {
		feedback.Errorf("Error during config init: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Info("Config file successfully initialized as: %s", file)
}
