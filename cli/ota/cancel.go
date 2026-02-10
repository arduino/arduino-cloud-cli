// This file is part of arduino-cloud-cli.
//
// Copyright (C) ARDUINO SRL (http://www.arduino.cc)
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

package ota

import (
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/ota"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/spf13/cobra"
)

type cancelFlags struct {
	otaID string
}

func initOtaCancelCommand() *cobra.Command {
	flags := &cancelFlags{}
	uploadCommand := &cobra.Command{
		Use:   "cancel",
		Short: "OTA cancel",
		Long:  "Cancel OTA by OTA ID",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runOtaCancelCommand(flags); err != nil {
				feedback.Errorf("Error during ota cancel: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	uploadCommand.Flags().StringVarP(&flags.otaID, "ota-id", "o", "", "OTA ID")

	return uploadCommand
}

func runOtaCancelCommand(flags *cancelFlags) error {
	if flags.otaID == "" {
		return fmt.Errorf("required flag \"ota-id\" not set")
	}

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	return ota.CancelOta(flags.otaID, cred)
}
