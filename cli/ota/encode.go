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

type encodeBinaryFlags struct {
	deviceID string
	file     string
}

func initEncodeBinaryCommand() *cobra.Command {
	flags := &encodeBinaryFlags{}
	uploadCommand := &cobra.Command{
		Use:   "encode",
		Short: "OTA firmware encode",
		Long:  "encode binary firmware to make it compatible with OTA",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runEncodeCommand(flags); err != nil {
				feedback.Errorf("Error during firmware encoding: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	uploadCommand.Flags().StringVarP(&flags.deviceID, "device-id", "d", "", "Device ID")
	uploadCommand.Flags().StringVarP(&flags.file, "file", "", "", "Binary file (.bin) to be encoded")
	uploadCommand.MarkFlagRequired("device-id")
	uploadCommand.MarkFlagRequired("file")
	return uploadCommand
}

func runEncodeCommand(flags *encodeBinaryFlags) error {
	logrus.Infof("Encoding binary %s", flags.file)

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	params := &ota.EncodeParams{
		DeviceID: flags.deviceID,
		File:     flags.file,
	}
	otafile, err := ota.Encode(context.TODO(), params, cred)
	if err != nil {
		return err
	}

	logrus.Info("Encode successfully performed. OTA file: ", *otafile)
	return nil
}
