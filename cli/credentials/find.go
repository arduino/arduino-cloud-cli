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
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initFindCommand() *cobra.Command {
	findCommand := &cobra.Command{
		Use:   "find",
		Short: "Find the credentials that would be used in your current directory",
		Long:  "Find the credentials to access Arduino IoT Cloud that would be used in your current directory",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runFindCommand(); err != nil {
				feedback.Errorf("Error during credentials find: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}

	return findCommand
}

func runFindCommand() error {
	logrus.Info("Looking for credentials")

	src, err := config.FindCredentials()
	if err != nil {
		return err
	}

	feedback.Printf("Using credentials in: %s", src)
	return nil
}
