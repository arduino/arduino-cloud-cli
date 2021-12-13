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

const (
	ClientIDLen     = 32
	ClientSecretLen = 64

	EnvPrefix = "ARDUINO_CLOUD"
)

// Config contains all the configuration parameters
// known by arduino-cloud-cli.
type Config struct {
	Client string `map-structure:"client"` // Client ID of the user
	Secret string `map-structure:"secret"` // Secret ID of the user, unique for each Client ID
}

// Validate the config
// If config is not valid, it returns an error explaining the reason
func (c *Config) Validate() error {
	if len(c.Client) != ClientIDLen {
		return fmt.Errorf(
			"client id not valid, expected len %d but got %d",
			ClientIDLen,
			len(c.Client),
		)
	}
	if len(c.Secret) != ClientSecretLen {
		return fmt.Errorf(
			"client secret not valid, expected len %d but got %d",
			ClientSecretLen,
			len(c.Secret),
		)
	}
	return nil
}

// IsEmpty checks if config has no params set
func (c *Config) IsEmpty() bool {
	if len(c.Client) != 0 {
		return false
	}
	if len(c.Secret) != 0 {
		return false
	}
	return true
}

// Retrieve looks for configuration parameters in
// environment variables or in configuration file
// Returns error if no config is found
func Retrieve() (*Config, error) {
	// Config extracted from environment has highest priority
	c, err := fromEnv()
	if err != nil {
		return nil, fmt.Errorf("reading config from environment variables: %w", err)
	}
	// Return the config only if it has been found
	if c != nil {
		return c, nil
	}

	c, err = fromFile()
	if err != nil {
		return nil, fmt.Errorf("reading config from file: %w", err)
	}
	if c != nil {
		return c, nil
	}

	return nil, fmt.Errorf(
		"config has not been found neither in environment variables " +
			"nor in the current directory, its parents or in arduino15",
	)
}

// fromFile looks for a configuration file
// If a config file is not found, it returns a nil config without raising errors.
// If invalid config file is found, it returns an error.
func fromFile() (*Config, error) {
	// Looks for a configuration file
	configDir, err := searchConfigDir()
	if err != nil {
		return nil, fmt.Errorf("can't get config directory: %w", err)
	}
	// Return nil config if no config file is found
	if configDir == nil {
		return nil, nil
	}

	v := viper.New()
	v.SetConfigName(Filename)
	v.AddConfigPath(*configDir)
	err = v.ReadInConfig()
	if err != nil {
		err = fmt.Errorf(
			"config file found at %s but cannot read its content: %w",
			*configDir,
			err,
		)
		return nil, err
	}

	conf := &Config{}
	err = v.Unmarshal(conf)
	if err != nil {
		return nil, fmt.Errorf(
			"config file found at %s but cannot unmarshal it: %w",
			*configDir,
			err,
		)
	}
	if err = conf.Validate(); err != nil {
		return nil, fmt.Errorf(
			"config file found at %s but is not valid: %w",
			*configDir,
			err,
		)
	}
	return conf, nil
}

// fromEnv looks for configuration credentials in environment variables.
// If credentials are not found, it returns a nil config without raising errors.
// If invalid credentials are found, it returns an error.
func fromEnv() (*Config, error) {
	v := viper.New()
	SetDefaults(v)
	v.SetEnvPrefix(EnvPrefix)
	v.AutomaticEnv()

	conf := &Config{}
	err := v.Unmarshal(conf)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal config from environment variables: %w", err)
	}

	if conf.IsEmpty() {
		return nil, nil
	}

	if err = conf.Validate(); err != nil {
		return nil, fmt.Errorf(
			"config retrieved from environment variables with prefix '%s' are not valid: %w",
			EnvPrefix,
			err,
		)
	}
	return conf, nil
}

// searchConfigDir configuration file in different directories in the following order:
// current working directory, parents of the current working directory, arduino15 default directory
// Returns a nil string if no config file has been found, without raising errors
// Returns an error if any problem is encountered during the file research which prevents
// to understand whether a config file exists or not
func searchConfigDir() (*string, error) {
	// Search in current directory and its parents.
	cwd, err := paths.Getwd()
	if err != nil {
		return nil, err
	}
	// Don't let bad naming mislead you, cwd.Parents()[0] is cwd itself so
	// we look in the current directory first and then on its parents.
	for _, path := range cwd.Parents() {
		if path.Join(Filename+".yaml").Exist() || path.Join(Filename+".json").Exist() {
			p := path.String()
			return &p, nil
		}
	}

	// Search in arduino's default data directory.
	arduino15, err := arduino.DataDir()
	if err != nil {
		return nil, err
	}
	if arduino15.Join(Filename+".yaml").Exist() || arduino15.Join(Filename+".json").Exist() {
		p := arduino15.String()
		return &p, nil
	}

	// Didn't find config file in the current directory, its parents or in arduino15"
	return nil, nil
}
