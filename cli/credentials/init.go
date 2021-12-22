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

package credentials

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/arduino"
	"github.com/arduino/arduino-cloud-cli/internal/config"
	"github.com/arduino/go-paths-helper"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initFlags struct {
	destDir   string
	overwrite bool
	format    string
}

func initInitCommand() *cobra.Command {
	initCommand := &cobra.Command{
		Use:   "init",
		Short: "Initialize a credentials file",
		Long:  "Initialize an Arduino IoT Cloud CLI credentials file",
		Run:   runInitCommand,
	}

	initCommand.Flags().StringVar(&initFlags.destDir, "dest-dir", "", "Sets where to save the credentials file")
	initCommand.Flags().BoolVar(&initFlags.overwrite, "overwrite", false, "Overwrite existing credentials file")
	initCommand.Flags().StringVar(&initFlags.format, "file-format", "yaml", "Format of the credentials file, can be {yaml|json}")

	return initCommand
}

func runInitCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Initializing credentials file")

	// Get default destination directory if it's not passed
	if initFlags.destDir == "" {
		credPath, err := arduino.DataDir()
		if err != nil {
			feedback.Errorf("Error during credentials init: cannot retrieve arduino default directory: %v", err)
			os.Exit(errorcodes.ErrGeneric)
		}
		// Create arduino default directory if it does not exist
		if credPath.NotExist() {
			if err = credPath.MkdirAll(); err != nil {
				feedback.Errorf("Error during credentials init: cannot create arduino default directory %s: %v", credPath, err)
				os.Exit(errorcodes.ErrGeneric)
			}
		}
		initFlags.destDir = credPath.String()
	}

	// Validate format flag
	initFlags.format = strings.ToLower(initFlags.format)
	if initFlags.format != "json" && initFlags.format != "yaml" {
		feedback.Error("Error during credentials init: format is not valid, provide 'json' or 'yaml'")
		os.Exit(errorcodes.ErrGeneric)
	}

	// Check that the destination directory is valid and build the credentials file path
	credPath, err := paths.New(initFlags.destDir).Abs()
	if err != nil {
		feedback.Errorf("Error during credentials init: cannot retrieve absolute path of %s: %v", initFlags.destDir, err)
		os.Exit(errorcodes.ErrGeneric)
	}
	if !credPath.IsDir() {
		feedback.Errorf("Error during credentials init: %s is not a valid directory", credPath)
		os.Exit(errorcodes.ErrGeneric)
	}
	credFile := credPath.Join(config.CredentialsFilename + "." + initFlags.format)
	if !initFlags.overwrite && credFile.Exist() {
		feedback.Errorf("Error during credentials init: %s already exists, use '--overwrite' to overwrite it",
			credFile)
		os.Exit(errorcodes.ErrGeneric)
	}

	// Take needed credentials starting an interactive mode
	feedback.Print("To obtain your API credentials visit https://create.arduino.cc/iot/integrations")
	id, key, err := paramsPrompt()
	if err != nil {
		feedback.Errorf("Error during credentials init: cannot take credentials params: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	// Write the credentials file
	newSettings := viper.New()
	newSettings.SetConfigPermissions(os.FileMode(0600))
	newSettings.Set("client", id)
	newSettings.Set("secret", key)
	if err := newSettings.WriteConfigAs(credFile.String()); err != nil {
		feedback.Errorf("Error during credentials init: cannot write credentials file: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.Printf("Credentials file successfully initialized at: %s", credFile)
}

func paramsPrompt() (id, key string, err error) {
	prompt := promptui.Prompt{
		Label: "Please enter the Client ID",
		Validate: func(s string) error {
			if len(s) != config.ClientIDLen {
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
			if len(s) != config.ClientSecretLen {
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
