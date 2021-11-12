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
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/ota"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var uploadFlags struct {
	deviceIDs []string
	tags      map[string]string
	file      string
	deferred  bool
	fqbn      string
}

func initUploadCommand() *cobra.Command {
	uploadCommand := &cobra.Command{
		Use:   "upload",
		Short: "OTA upload",
		Long:  "OTA upload on a device of Arduino IoT Cloud",
		Run:   runUploadCommand,
	}

	uploadCommand.Flags().StringSliceVarP(&uploadFlags.deviceIDs, "device-ids", "d", []string{},
		"Comma-separated list of device IDs to update")
	uploadCommand.Flags().StringToStringVar(&uploadFlags.tags, "tags", nil,
		"Comma-separated list of tags with format <key>=<value>.\n"+
			"Perform and OTA upload on all devices that match the provided tags.\n"+
			"Mutually exclusive with `--device-id`.",
	)
	uploadCommand.Flags().StringVarP(&uploadFlags.file, "file", "", "", "Binary file (.bin) to be uploaded")
	uploadCommand.Flags().BoolVar(&uploadFlags.deferred, "deferred", false, "Perform a deferred OTA. It can take up to 1 week.")
	uploadCommand.Flags().StringVarP(&uploadFlags.fqbn, "fqbn", "b", "", "FQBN of the devices to update")

	uploadCommand.MarkFlagRequired("file")
	uploadCommand.MarkFlagRequired("fqbn")
	return uploadCommand
}

func runUploadCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Uploading binary %s", uploadFlags.file)

	params := &ota.UploadParams{
		DeviceIDs: uploadFlags.deviceIDs,
		Tags:      uploadFlags.tags,
		File:      uploadFlags.file,
		Deferred:  uploadFlags.deferred,
		FQBN:      uploadFlags.fqbn,
	}

	resp, err := ota.Upload(params)
	if err != nil {
		feedback.Errorf("Error during ota upload: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	devs := strings.Join(resp.Updated, ",")
	devs = strings.TrimRight(devs, ",")
	success := fmt.Sprintf("Successfully sent OTA request to: %s", devs)

	devs = strings.Join(resp.Invalid, ",")
	devs = strings.TrimRight(devs, ",")
	invalid := fmt.Sprintf("Cannot send OTA request to: %s", devs)

	devs = strings.Join(resp.Failed, ",")
	devs = strings.TrimRight(devs, ",")
	fail := fmt.Sprintf("Failed to send OTA request to: %s", devs)

	det := strings.Join(resp.Errors, "\n")
	det = strings.TrimRight(det, ",")
	details := fmt.Sprintf("\nDetails:\n%s", det)

	feedback.Printf(success, invalid, fail, details)
	logrus.Info("Upload successfully started")
}
