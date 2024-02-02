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
	"github.com/spf13/cobra"
)

type decodeHeaderFlags struct {
	file string
}

func initDecodeHeaderCommand() *cobra.Command {
	flags := &decodeHeaderFlags{}
	uploadCommand := &cobra.Command{
		Use:   "decode",
		Short: "OTA firmware header decoder",
		Long:  "decode OTA firmware header of the given binary file",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runDecodeHeaderCommand(flags); err != nil {
				feedback.Errorf("Error during firmware decoding: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	uploadCommand.Flags().StringVarP(&flags.file, "file", "", "", "Binary file (.ota)")
	uploadCommand.MarkFlagRequired("file")
	return uploadCommand
}

func runDecodeHeaderCommand(flags *decodeHeaderFlags) error {
	params := &ota.ReadHeaderParams{
		File: flags.file,
	}
	err := ota.ReadHeader(params)
	if err != nil {
		return err
	}
	return nil
}
