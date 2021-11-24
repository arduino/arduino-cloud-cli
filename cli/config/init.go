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
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/internal/config"
	"github.com/arduino/go-paths-helper"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	clientIDLen     = 32
	clientSecretLen = 64
)

var initFlags struct {
	destDir   string
	overwrite bool
	format    string
}

func initInitCommand() *cobra.Command {
	initCommand := &cobra.Command{
		Use:   "init",
		Short: "Initialize a configuration file",
		Long:  "Initialize an Arduino IoT Cloud CLI configuration",
		Run:   runInitCommand,
	}

	initCommand.Flags().StringVar(&initFlags.destDir, "dest-dir", "", "Sets where to save the configuration file")
	initCommand.Flags().BoolVar(&initFlags.overwrite, "overwrite", false, "Overwrite existing config file")
	initCommand.Flags().StringVar(&initFlags.format, "config-format", "yaml", "Format of the configuration file, can be {yaml|json}")

	return initCommand
}

func runInitCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Initializing config file")

	// Get default destination directory if it's not passed
	if initFlags.destDir == "" {
		configPath, err := config.ArduinoDir()
		initFlags.destDir = configPath.String()
		if err != nil {
			feedback.Errorf("Error during config init: cannot retrieve arduino default directory: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}
	}

	// Validate format flag
	initFlags.format = strings.ToLower(initFlags.format)
	if initFlags.format != "json" && initFlags.format != "yaml" {
		feedback.Error("Error during config init: passed format is not valid, select between 'json' and 'yaml'")
		os.Exit(errorcodes.ErrGeneric)
	}

	// Check that the destination directory is valid and build the configuration file path
	configPath, err := paths.New(initFlags.destDir).Abs()
	if err != nil {
		feedback.Errorf("Error during config init: cannot retrieve absolute path of passed dest-dir: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
	if !configPath.IsDir() {
		feedback.Error("Error during config init: passed dest-dir is not a valid directory")
		os.Exit(errorcodes.ErrGeneric)
	}
	configFile := configPath.Join(config.Filename + "." + initFlags.format)
	if !initFlags.overwrite && configFile.Exist() {
		feedback.Error("Error during config init: config file already exists, use --overwrite to discard the existing one")
		os.Exit(errorcodes.ErrGeneric)
	}

	// Take needed configuration parameters starting an interactive mode
	feedback.Print("To obtain your API credentials visit https://create.arduino.cc/iot/integrations")
	id, key, err := paramsPrompt()
	if err != nil {
		feedback.Errorf("Error during config init: taking config parameters: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	// Write the configuration file
	newSettings := viper.New()
	newSettings.SetConfigPermissions(os.FileMode(0600))
	newSettings.Set("client", id)
	newSettings.Set("secret", key)
	if err := newSettings.WriteConfigAs(configFile.String()); err != nil {
		feedback.Errorf("Error during config init: cannot create config file: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Infof("Config file successfully initialized at: %s", configFile.String())
}

func paramsPrompt() (id, key string, err error) {
	prompt := promptui.Prompt{
		Label: "Please enter the Client ID",
		Validate: func(s string) error {
			if len(s) != clientIDLen {
				return errors.New("client-id not valid")
			}
			return nil
		},
	}
	id, err = prompt.Run()
	if err != nil {
		return "", "", fmt.Errorf("client prompt fail: %w", err)
	}

	prompt = promptui.Prompt{
		Mask:  '*',
		Label: "Please enter the Client Secret",
		Validate: func(s string) error {
			if len(s) != clientSecretLen {
				return errors.New("client secret not valid")
			}
			return nil
		},
	}
	key, err = prompt.Run()
	if err != nil {
		return "", "", fmt.Errorf("client secret prompt fail: %w", err)
	}

	return id, key, nil
}
