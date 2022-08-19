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
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/go-paths-helper"
	"github.com/manifoldco/promptui"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type initFlags struct {
	destDir   string
	overwrite bool
	format    string
}

func initInitCommand() *cobra.Command {
	flags := &initFlags{}
	initCommand := &cobra.Command{
		Use:   "init",
		Short: "Initialize a credentials file",
		Long:  "Initialize an Arduino IoT Cloud CLI credentials file",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runInitCommand(flags); err != nil {
				feedback.Errorf("Error during credentials init: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}

	initCommand.Flags().StringVar(&flags.destDir, "dest-dir", "", "Sets where to save the credentials file")
	initCommand.Flags().BoolVar(&flags.overwrite, "overwrite", false, "Overwrite existing credentials file")
	initCommand.Flags().StringVar(&flags.format, "file-format", "yaml", "Format of the credentials file, can be {yaml|json}")

	return initCommand
}

func runInitCommand(flags *initFlags) error {
	logrus.Info("Initializing credentials file")

	// Get default destination directory if it's not passed
	if flags.destDir == "" {
		credPath, err := arduino.DataDir()
		if err != nil {
			return fmt.Errorf("cannot retrieve arduino default directory: %w", err)
		}
		// Create arduino default directory if it does not exist
		if credPath.NotExist() {
			if err = credPath.MkdirAll(); err != nil {
				return fmt.Errorf("cannot create arduino default directory %s: %w", credPath, err)
			}
		}
		flags.destDir = credPath.String()
	}

	// Validate format flag
	flags.format = strings.ToLower(flags.format)
	if flags.format != "json" && flags.format != "yaml" {
		return fmt.Errorf("format is not valid, provide 'json' or 'yaml'")
	}

	// Check that the destination directory is valid and build the credentials file path
	credPath, err := paths.New(flags.destDir).Abs()
	if err != nil {
		return fmt.Errorf("cannot retrieve absolute path of %s: %w", flags.destDir, err)
	}
	if !credPath.IsDir() {
		return fmt.Errorf("%s is not a valid directory", credPath)
	}
	credFile := credPath.Join(config.CredentialsFilename + "." + flags.format)
	if !flags.overwrite && credFile.Exist() {
		return fmt.Errorf("%s already exists, use '--overwrite' to overwrite it", credFile)
	}

	// Take needed credentials starting an interactive mode
	feedback.Print("To obtain your API credentials visit https://create.arduino.cc/iot/integrations")
	id, key, org, err := paramsPrompt()
	if err != nil {
		return fmt.Errorf("cannot take credentials params: %w", err)
	}

	// Write the credentials file
	newSettings := viper.New()
	newSettings.SetConfigPermissions(os.FileMode(0600))
	newSettings.Set("client", id)
	newSettings.Set("secret", key)
	newSettings.Set("organization", org)
	if err := newSettings.WriteConfigAs(credFile.String()); err != nil {
		return fmt.Errorf("cannot write credentials file: %w", err)
	}

	feedback.Printf("Credentials file successfully initialized at: %s", credFile)
	return nil
}

func paramsPrompt() (id, key, org string, err error) {
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
		return "", "", "", fmt.Errorf("client prompt fail: %w", err)
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
		return "", "", "", fmt.Errorf("client secret prompt fail: %w", err)
	}

	prompt = promptui.Prompt{
		Mask:  '*',
		Label: "Please enter the Organization ID - if any - Leave empty otherwise",
		Validate: func(s string) error {
			if len(s) != 0 && len(s) != config.OrganizationLen {
				return errors.New("organization id not valid")
			}
			return nil
		},
	}
	org, err = prompt.Run()
	if err != nil {
		return "", "", "", fmt.Errorf("organization id prompt fail: %w", err)
	}

	return id, key, org, nil
}
