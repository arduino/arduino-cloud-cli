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
	"os"

	"github.com/spf13/viper"
)

// Config contains all the configuration parameters
// known by arduino-cloud-cli
type Config struct {
	Client string `map-structure:"client"` // Client ID of the user
	Secret string `map-structure:"secret"` // Secret ID of the user, unique for each Client ID
}

// Retrieve returns the actual parameters contained in the
// configuration file, if any. Returns error if no config file is found.
func Retrieve() (*Config, error) {
	conf := &Config{}
	v := viper.New()
	v.SetConfigName(Filename)
	v.AddConfigPath(".")
	err := v.ReadInConfig()
	if err != nil {
		fmt.Errorf("%s: %w", "retrieving config file", err)
	} else {
		v.Unmarshal(conf)
	}

	client, found := os.LookupEnv("ARDUINO_CLOUD_CLIENT")
	if !found {
		err = fmt.Errorf("%s: %w", "Unable to retrieve token client", err)
		return nil, err

	}
	secret, found := os.LookupEnv("ARDUINO_CLOUD_SECRET")
	if !found {
		err = fmt.Errorf("%s: %w", "Unable to retrieve token secret", err)
		return nil, err
	}

	conf.Client = client
	conf.Secret = secret
	return conf, nil
}
