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

package version

import (
	"os"

	"github.com/arduino/arduino-cli/cli/feedback"
	v "github.com/arduino/arduino-cloud-cli/version"
	"github.com/spf13/cobra"
)

// NewCommand creates a new `version` command.
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Shows version number of Arduino Cloud CLI.",
		Long:    "Shows the version number of Arduino Cloud CLI which is installed on your system.",
		Example: "  " + os.Args[0] + " version",
		Args:    cobra.NoArgs,
		Run:     run,
	}
}

func run(cmd *cobra.Command, args []string) {
	feedback.Print(v.VersionInfo)
}
