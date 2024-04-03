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

package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/cli/credentials"
	"github.com/arduino/arduino-cloud-cli/cli/dashboard"
	"github.com/arduino/arduino-cloud-cli/cli/device"
	"github.com/arduino/arduino-cloud-cli/cli/ota"
	"github.com/arduino/arduino-cloud-cli/cli/thing"
	"github.com/arduino/arduino-cloud-cli/cli/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type cliFlags struct {
	verbose      bool
	outputFormat string
}

func Execute() {
	flags := &cliFlags{}
	cli := &cobra.Command{
		Use:   "arduino-cloud-cli",
		Short: "Arduino Cloud CLI.",
		Long:  "Arduino Cloud Command Line Interface (arduino-cloud-cli).",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if err := preRun(flags); err != nil {
				feedback.Error(err)
				os.Exit(errorcodes.ErrBadCall)
			}
		},
	}
	cli.PersistentFlags().BoolVarP(&flags.verbose, "verbose", "v", false, "Print the logs on the standard output.")
	validOutputFormats := []string{"text", "json", "jsonmini", "yaml"}
	cli.PersistentFlags().StringVar(&flags.outputFormat, "format", "text",
		fmt.Sprintf("The output format, can be: %s", strings.Join(validOutputFormats, ", ")),
	)

	cli.AddCommand(version.NewCommand())
	cli.AddCommand(credentials.NewCommand())
	cli.AddCommand(device.NewCommand())
	cli.AddCommand(thing.NewCommand())
	cli.AddCommand(dashboard.NewCommand())
	cli.AddCommand(ota.NewCommand())

	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func parseFormatString(arg string) (feedback.OutputFormat, bool) {
	f, found := map[string]feedback.OutputFormat{
		"json":     feedback.JSON,
		"jsonmini": feedback.JSONMini,
		"text":     feedback.Text,
		"yaml":     feedback.YAML,
	}[strings.ToLower(arg)]

	return f, found
}

func preRun(flags *cliFlags) error {
	logrus.SetOutput(io.Discard)
	// enable log only if verbose flag is passed
	if flags.verbose {
		logrus.SetLevel(logrus.InfoLevel)
		logrus.SetOutput(os.Stdout)
	}

	// normalize the format strings
	flags.outputFormat = strings.ToLower(flags.outputFormat)
	// check the right output format was passed
	format, found := parseFormatString(flags.outputFormat)
	if !found {
		return fmt.Errorf("invalid output format: %s", flags.outputFormat)
	}
	// use the output format to configure the Feedback
	feedback.SetFormat(format)
	return nil
}
