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
	"fmt"

	"github.com/arduino/arduino-cloud-cli/arduino"
	"github.com/arduino/go-paths-helper"
	"github.com/spf13/viper"
)

// Config contains all the configuration parameters
// known by arduino-cloud-cli.
type Config struct {
	Client string `map-structure:"client"` // Client ID of the user
	Secret string `map-structure:"secret"` // Secret ID of the user, unique for each Client ID
}

// Retrieve returns the actual parameters contained in the
// configuration file, if any. Returns error if no config file is found.
func Retrieve() (*Config, error) {
	configDir, err := searchConfigDir()
	if err != nil {
		return nil, fmt.Errorf("can't get config directory: %w", err)
	}

	v := viper.New()
	v.SetConfigName(Filename)
	v.AddConfigPath(configDir)
	err = v.ReadInConfig()
	if err != nil {
		err = fmt.Errorf("%s: %w", "retrieving config file", err)
		return nil, err
	}

	conf := &Config{}
	v.Unmarshal(conf)
	return conf, nil
}

func searchConfigDir() (string, error) {
	// Search in current directory and its parents.
	cwd, err := paths.Getwd()
	if err != nil {
		return "", err
	}
	// Don't let bad naming mislead you, cwd.Parents()[0] is cwd itself so
	// we look in the current directory first and then on its parents.
	for _, path := range cwd.Parents() {
		if path.Join(Filename+".yaml").Exist() || path.Join(Filename+".json").Exist() {
			return path.String(), nil
		}
	}

	// Search in arduino's default data directory.
	arduino15, err := arduino.DataDir()
	if err != nil {
		return "", err
	}
	if arduino15.Join(Filename+".yaml").Exist() || arduino15.Join(Filename+".json").Exist() {
		return arduino15.String(), nil
	}

	return "", fmt.Errorf(
		"didn't find config file in the current directory, its parents or in %s",
		arduino15.String(),
	)
}
