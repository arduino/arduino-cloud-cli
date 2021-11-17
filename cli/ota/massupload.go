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
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/ota"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var massUploadFlags struct {
	deviceIDs []string
	tags      map[string]string
	file      string
	deferred  bool
	fqbn      string
}

func initMassUploadCommand() *cobra.Command {
	massUploadCommand := &cobra.Command{
		Use:   "mass-upload",
		Short: "Mass OTA upload",
		Long:  "Mass OTA upload on devices of Arduino IoT Cloud",
		Run:   runMassUploadCommand,
	}

	massUploadCommand.Flags().StringSliceVarP(&massUploadFlags.deviceIDs, "device-ids", "d", nil,
		"Comma-separated list of device IDs to update")
	massUploadCommand.Flags().StringToStringVar(&massUploadFlags.tags, "device-tags", nil,
		"Comma-separated list of tags with format <key>=<value>.\n"+
			"Perform an OTA upload on all devices that match the provided tags.\n"+
			"Mutually exclusive with `--device-ids`.",
	)
	massUploadCommand.Flags().StringVarP(&massUploadFlags.file, "file", "", "", "Binary file (.bin) to be uploaded")
	massUploadCommand.Flags().BoolVar(&massUploadFlags.deferred, "deferred", false, "Perform a deferred OTA. It can take up to 1 week.")
	massUploadCommand.Flags().StringVarP(&massUploadFlags.fqbn, "fqbn", "b", "", "FQBN of the devices to update")

	massUploadCommand.MarkFlagRequired("file")
	massUploadCommand.MarkFlagRequired("fqbn")
	return massUploadCommand
}

func runMassUploadCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Uploading binary %s", massUploadFlags.file)

	params := &ota.MassUploadParams{
		DeviceIDs: massUploadFlags.deviceIDs,
		Tags:      massUploadFlags.tags,
		File:      massUploadFlags.file,
		Deferred:  massUploadFlags.deferred,
		FQBN:      massUploadFlags.fqbn,
	}

	resp, err := ota.MassUpload(params)
	if err != nil {
		feedback.Errorf("Error during ota upload: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	success := strings.Join(resp.Updated, ",")
	success = strings.TrimRight(success, ",")
	feedback.Printf("\nSuccessfully sent OTA request to: %s", success)

	invalid := strings.Join(resp.Invalid, ",")
	invalid = strings.TrimRight(invalid, ",")
	feedback.Printf("Cannot send OTA request to: %s", invalid)

	fail := strings.Join(resp.Failed, ",")
	fail = strings.TrimRight(fail, ",")
	feedback.Printf("Failed to send OTA request to: %s", fail)

	det := strings.Join(resp.Errors, "\n")
	feedback.Printf("\nDetails:\n%s", det)
}
