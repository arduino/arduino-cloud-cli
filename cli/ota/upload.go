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

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/ota"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var uploadFlags struct {
	deviceID string
	file     string
	deferred bool
}

func initUploadCommand() *cobra.Command {
	uploadCommand := &cobra.Command{
		Use:   "upload",
		Short: "OTA upload",
		Long:  "OTA upload on a device of Arduino IoT Cloud",
		Run:   runUploadCommand,
	}

	uploadCommand.Flags().StringVarP(&uploadFlags.deviceID, "device-id", "d", "", "Device ID")
	uploadCommand.Flags().StringVarP(&uploadFlags.file, "file", "", "", "Binary file (.bin) to be uploaded")
	uploadCommand.Flags().BoolVar(&uploadFlags.deferred, "deferred", false, "Perform a deferred OTA. It can take up to 1 week.")

	uploadCommand.MarkFlagRequired("device-id")
	uploadCommand.MarkFlagRequired("file")
	return uploadCommand
}

func runUploadCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Uploading binary %s to device %s", uploadFlags.file, uploadFlags.deviceID)

	params := &ota.UploadParams{
		DeviceID: uploadFlags.deviceID,
		File:     uploadFlags.file,
	}
	err := ota.Upload(params)
	if err != nil {
		feedback.Errorf("Error during ota upload: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Info("Upload successfully started")
}
