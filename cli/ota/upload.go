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
	"context"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/ota"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type uploadFlags struct {
	deviceID string
	file     string
	deferred bool
}

func initUploadCommand() *cobra.Command {
	flags := &uploadFlags{}
	uploadCommand := &cobra.Command{
		Use:   "upload",
		Short: "OTA upload",
		Long:  "OTA upload on a device of Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runUploadCommand(flags); err != nil {
				feedback.Errorf("Error during ota upload: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	uploadCommand.Flags().StringVarP(&flags.deviceID, "device-id", "d", "", "Device ID")
	uploadCommand.Flags().StringVarP(&flags.file, "file", "", "", "Binary file (.bin) to be uploaded")
	uploadCommand.Flags().BoolVar(&flags.deferred, "deferred", false, "Perform a deferred OTA. It can take up to 1 week.")
	uploadCommand.MarkFlagRequired("device-id")
	uploadCommand.MarkFlagRequired("file")
	return uploadCommand
}

func runUploadCommand(flags *uploadFlags) error {
	logrus.Infof("Uploading binary %s to device %s", flags.file, flags.deviceID)

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	params := &ota.UploadParams{
		DeviceID: flags.deviceID,
		File:     flags.file,
		Deferred: flags.deferred,
	}
	err = ota.Upload(context.TODO(), params, cred)
	if err != nil {
		return err
	}

	logrus.Info("Upload successfully started")
	return nil
}
