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
		Short: "Find the credentials file being used in your current directory",
		Long:  "Find the Arduino Cloud CLI credentials file being used in your current directory",
		Run:   runFindCommand,
	}

	return findCommand
}

func runFindCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Looking for a credentials file")

	src, err := config.FindCredentials()
	if err != nil {
		feedback.Error("Error during credentials find: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.Printf("Using credentials in: %s", src)
}
