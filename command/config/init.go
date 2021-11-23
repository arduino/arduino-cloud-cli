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

	"github.com/arduino/arduino-cloud-cli/internal/config"
	"github.com/arduino/go-paths-helper"
	"github.com/manifoldco/promptui"
	"github.com/spf13/viper"
)

const (
	clientLen = 32
	secretLen = 64
)

// InitParams contains the parameters needed to initialize a configuration file.
type InitParams struct {
	DestDir   string // Destination directory in which the configuration file will be saved
	Overwrite bool   // Overwrite specifies if existing config file should be overwritten
	Format    string // Config file format, can be 'json' or 'yaml'
}

func validateFormatString(arg string) error {
	if arg != "json" && arg != "yaml" {
		return errors.New("passed format is not valid, select between 'json' and 'yaml'")
	}
	return nil
}

// Init initializes a configuration file with values passed in by the user.
// If the file doesn't exist, it is created.
// If it already exists, it is written to only if overwrite param is true.
func Init(params *InitParams) (filepath string, err error) {
	var configPath *paths.Path
	if params.DestDir != "" {
		configPath, err = paths.New(params.DestDir).Abs()
		if err != nil {
			return "", fmt.Errorf("cannot retrieve absolute path of passed dest-dir: %w", err)
		}
	} else {
		configPath, err = config.ArduinoDir()
		if err != nil {
			return "", fmt.Errorf("cannot retrieve arduino default directory: %w", err)
		}
	}

	if !configPath.IsDir() {
		return "", fmt.Errorf("chosen destination dir is not valid: %w", err)
	}

	params.Format = strings.ToLower(params.Format)
	if err := validateFormatString(params.Format); err != nil {
		return "", err
	}

	configFile := configPath.Join(config.Filename + "." + params.Format)
	if !params.Overwrite && configFile.Exist() {
		return "", errors.New("config file already exists, use --overwrite to discard the existing one")
	}

	id, key, err := prompt()
	if err != nil {
		return "", err
	}

	newSettings := viper.New()
	newSettings.SetConfigPermissions(os.FileMode(0600))
	newSettings.Set("client", id)
	newSettings.Set("secret", key)
	if err := newSettings.WriteConfigAs(configFile.String()); err != nil {
		return "", fmt.Errorf("cannot create config file: %v", err)
	}

	return configFile.String(), nil
}

func prompt() (id, key string, err error) {
	prompt := promptui.Prompt{
		Label: "client",
		Validate: func(s string) error {
			if len(s) != clientLen {
				return errors.New("")
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
		Label: "secret",
		Validate: func(s string) error {
			if len(s) != secretLen {
				return errors.New("")
			}
			return nil
		},
	}
	key, err = prompt.Run()
	if err != nil {
		return "", "", fmt.Errorf("secret prompt fail: %w", err)
	}

	return id, key, nil
}
