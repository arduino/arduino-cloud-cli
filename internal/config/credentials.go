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
	"strings"

	"github.com/arduino/go-paths-helper"
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

// SetEmptyCredentials sets the default credentials values to empty strings.
func SetEmptyCredentials(settings *viper.Viper) {
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

// FindCredentials looks for credentials in
// environment variables or in credentials file.
// Returns the source of found credentials (env or filepath).
// Returns an error if credentials are not found
// specifying paths where the credentials are searched.
func FindCredentials() (source string, err error) {
	// Credentials extracted from environment has highest priority
	logrus.Info("Looking for credentials in environment variables")
	c, err := fromEnv()
	if err != nil {
		return "", fmt.Errorf("looking for credentials in environment variables: %w", err)
	}
	if !c.IsEmpty() {
		logrus.Infof("Credentials found in environment variables with prefix '%s'", EnvPrefix)
		return "environment variables", nil
	}

	logrus.Info("Looking for credentials in file system")
	path, found, err := searchConfigFile(CredentialsFilename)
	if err != nil {
		return "", fmt.Errorf("looking for credentials files: %w", err)
	}
	if found {
		return path, nil
	}

	return "", fmt.Errorf(
		"credentials have not been found neither in environment variables " +
			"nor in the current directory, its parents or in arduino15",
	)
}

// RetrieveCredentials retrieves credentials from
// environment variables or credentials file.
// Returns error if credentials are not found or
// if found credentials are invalid.
func RetrieveCredentials() (cred *Credentials, err error) {
	// Credentials extracted from environment has highest priority
	logrus.Info("Looking for credentials in environment variables")
	cred, err = fromEnv()
	if err != nil {
		return nil, fmt.Errorf("reading credentials from environment variables: %w", err)
	}
	// Returns credentials if found in env
	if !cred.IsEmpty() {
		// Returns error if credentials are found but are not valid
		if err := cred.Validate(); err != nil {
			return nil, fmt.Errorf(
				"credentials retrieved from environment variables with prefix '%s' are not valid: %w", EnvPrefix, err,
			)
		}
		logrus.Infof("Credentials found in environment variables with prefix '%s'", EnvPrefix)
		return cred, nil
	}

	logrus.Info("Looking for credentials in file system")
	filepath, found, err := searchConfigFile(CredentialsFilename)
	if err != nil {
		return nil, fmt.Errorf("can't get credentials directory: %w", err)
	}
	// Returns credentials if found in a file
	if found {
		if cred, err = fromFile(filepath); err != nil {
			return nil, fmt.Errorf("reading credentials from file %s: %w", filepath, err)
		}
		// Returns error if credentials are found but are not valid
		if err := cred.Validate(); err != nil {
			return nil, fmt.Errorf(
				"credentials retrieved from file %s are not valid: %w", filepath, err,
			)
		}
		return cred, nil
	}

	return nil, fmt.Errorf(
		"credentials have not been found neither in environment variables " +
			"nor in the current directory, its parents or in arduino15",
	)
}

// fromFile retrieves credentials from a credentials file.
// Returns error if credentials are not found or cannot be fetched.
func fromFile(filepath string) (*Credentials, error) {
	v := viper.New()
	v.SetConfigFile(filepath)
	v.SetConfigType(strings.TrimLeft(paths.New(filepath).Ext(), "."))
	err := v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("cannot read credentials file: %w", err)
	}

	cred := &Credentials{}
	err = v.Unmarshal(cred)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal credentials file: %w", err)
	}
	return cred, nil
}

// fromEnv retrieves credentials from environment variables.
// Returns empty credentials if not found.
func fromEnv() (*Credentials, error) {
	v := viper.New()
	SetEmptyCredentials(v)
	v.SetEnvPrefix(EnvPrefix)
	v.AutomaticEnv()

	cred := &Credentials{}
	err := v.Unmarshal(cred)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal credentials from environment variables: %w", err)
	}
	return cred, nil
}
