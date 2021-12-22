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

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	// ClientIDLen specifies the length of Arduino IoT Cloud client ids.
	ClientIDLen = 32
	// ClientSecretLen specifies the length of Arduino IoT Cloud client secrets.
	ClientSecretLen = 64

	// EnvPrefix is the prefix environment variables should have to be
	// fetched as credentials parameters during the credentials retrieval.
	EnvPrefix = "ARDUINO_CLOUD"

	// CredentialsFilename specifies the name of the credentials file.
	CredentialsFilename = "arduino-cloud-credentials"
)

// SetDefaultCredentials sets the default credentials values.
func SetDefaultCredentials(settings *viper.Viper) {
	// Client ID
	settings.SetDefault("client", "")
	// Secret
	settings.SetDefault("secret", "")
}

// Credentials contains the parameters of Arduino IoT Cloud credentials.
type Credentials struct {
	Client string `map-structure:"client"` // Client ID of the user
	Secret string `map-structure:"secret"` // Secret ID of the user, unique for each Client ID
}

// Validate the credentials.
// If credentials are not valid, it returns an error explaining the reason.
func (c *Credentials) Validate() error {
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

// IsEmpty checks if credentials has no params set.
func (c *Credentials) IsEmpty() bool {
	return len(c.Client) == 0 && len(c.Secret) == 0
}

// RetrieveCredentials looks for credentials in
// environment variables or in credentials file.
// Returns error if no credentials are found.
func RetrieveCredentials() (*Credentials, error) {
	// Credentials extracted from environment has highest priority
	logrus.Info("Looking for credentials in environment variables")
	c, err := fromEnv()
	if err != nil {
		return nil, fmt.Errorf("reading credentials from environment variables: %w", err)
	}
	// Return credentials only if found
	if c != nil {
		logrus.Info("Credentials found in environment variables")
		return c, nil
	}

	logrus.Info("Looking for credentials in file system")
	c, err = fromFile()
	if err != nil {
		return nil, fmt.Errorf("reading credentials from file: %w", err)
	}
	if c != nil {
		return c, nil
	}

	return nil, fmt.Errorf(
		"credentials have not been found neither in environment variables " +
			"nor in the current directory, its parents or in arduino15",
	)
}

// fromFile looks for a credentials file.
// If a credentials file is not found, it returns nil credentials without raising errors.
// If invalid credentials file is found, it returns an error.
func fromFile() (*Credentials, error) {
	// Looks for a credentials file
	configDir, err := searchConfigDir(CredentialsFilename)
	if err != nil {
		return nil, fmt.Errorf("can't get credentials directory: %w", err)
	}
	// Return nil credentials if no config file is found
	if configDir == nil {
		return nil, nil
	}

	v := viper.New()
	v.SetConfigName(CredentialsFilename)
	v.AddConfigPath(*configDir)
	err = v.ReadInConfig()
	if err != nil {
		err = fmt.Errorf(
			"credentials file found at %s but cannot read its content: %w",
			*configDir,
			err,
		)
		return nil, err
	}

	cred := &Credentials{}
	err = v.Unmarshal(cred)
	if err != nil {
		return nil, fmt.Errorf(
			"credentials file found at %s but cannot unmarshal it: %w",
			*configDir,
			err,
		)
	}
	if err = cred.Validate(); err != nil {
		return nil, fmt.Errorf(
			"credentials file found at %s but is not valid: %w",
			*configDir,
			err,
		)
	}
	return cred, nil
}

// fromEnv looks for credentials in environment variables.
// If credentials are not found, it returns nil credentials without raising errors.
// If invalid credentials are found, it returns an error.
func fromEnv() (*Credentials, error) {
	v := viper.New()
	SetDefaultCredentials(v)
	v.SetEnvPrefix(EnvPrefix)
	v.AutomaticEnv()

	cred := &Credentials{}
	err := v.Unmarshal(cred)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal credentials from environment variables: %w", err)
	}

	if cred.IsEmpty() {
		return nil, nil
	}

	if err = cred.Validate(); err != nil {
		return nil, fmt.Errorf(
			"credentials retrieved from environment variables with prefix '%s' are not valid: %w",
			EnvPrefix,
			err,
		)
	}
	return cred, nil
}
