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

type statusFlags struct {
	otaID    string
	otaIDs   string
	deviceId string
	limit    int16
	sort     string
}

func initOtaStatusCommand() *cobra.Command {
	flags := &statusFlags{}
	statusCommand := &cobra.Command{
		Use:   "status",
		Short: "OTA status",
		Long:  "Get OTA status by OTA or device ID",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runPrintOtaStatusCommand(flags, cmd); err != nil {
				feedback.Errorf("\nError during ota get status: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	statusCommand.Flags().StringVarP(&flags.otaID, "ota-id", "o", "", "OTA ID")
	statusCommand.Flags().StringVarP(&flags.otaIDs, "ota-ids", "", "", "OTA IDs (comma separated)")
	statusCommand.Flags().StringVarP(&flags.deviceId, "device-id", "d", "", "Device ID")
	statusCommand.Flags().Int16VarP(&flags.limit, "limit", "l", 10, "Output limit (default: 10)")
	statusCommand.Flags().StringVarP(&flags.sort, "sort", "s", "desc", "Sorting (default: desc)")

	return statusCommand
}

func runPrintOtaStatusCommand(flags *statusFlags, command *cobra.Command) error {
	if flags.otaID == "" && flags.deviceId == "" && flags.otaIDs == "" {
		command.Help()
		return fmt.Errorf("required flag(s) \"ota-id\" or \"device-id\" or \"ota-ids\" not set")
	}

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	return ota.PrintOtaStatus(flags.otaID, flags.otaIDs, flags.deviceId, cred, int(flags.limit), flags.sort)
}
