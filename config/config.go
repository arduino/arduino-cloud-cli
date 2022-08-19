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
	"github.com/arduino/arduino-cloud-cli/arduino"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// searchConfigFile looks for a configuration file in different directories in the following order:
// current working directory, parents of the current working directory, arduino15 default directory.
// Returns empty string and false if no config file has been found, without raising errors.
// Returns an error if any problem is encountered during the file research which prevents
// to understand whether a config file exists or not.
func searchConfigFile(confname string) (dir string, found bool, err error) {
	// Search in current directory and its parents.
	cwd, err := paths.Getwd()
	if err != nil {
		return "", false, err
	}
	// Don't let bad naming mislead you, cwd.Parents()[0] is cwd itself so
	// we look in the current directory first and then on its parents.
	for _, path := range cwd.Parents() {
		logrus.Infof("Looking for %s in %s", confname, path)
		if file, found := configFileInDir(confname, path); found {
			logrus.Infof("Found %s at %s", confname, file)
			return file.String(), true, nil
		}
	}

	// Search in arduino's default data directory.
	arduino15, err := arduino.DataDir()
	if err != nil {
		return "", false, err
	}
	logrus.Infof("Looking for %s in %s", confname, arduino15)
	if file, found := configFileInDir(confname, arduino15); found {
		logrus.Infof("%s found at %s", confname, file)
		return file.String(), true, nil
	}

	// Didn't find config file in the current directory, its parents or in arduino15"
	return "", false, nil
}

// configFileInDir looks for a configuration file in the passed directory.
// If a configuration file is found, then it is returned.
// In case of multiple config files, it returns the one with the highest priority
// according to viper.
func configFileInDir(confname string, dir *paths.Path) (filepath *paths.Path, found bool) {
	for _, ext := range viper.SupportedExts {
		if filepath = dir.Join(confname + "." + ext); filepath.Exist() {
			return filepath, true
		}
	}
	return nil, false
}
