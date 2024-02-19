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
	"github.com/spf13/cobra"
)

type encodeBinaryFlags struct {
	FQBN string
	file string
}

func initEncodeBinaryCommand() *cobra.Command {
	flags := &encodeBinaryFlags{}
	uploadCommand := &cobra.Command{
		Use:   "encode",
		Short: "OTA firmware encoder",
		Long:  "encode header firmware to make it compatible with Arduino OTA",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runEncodeCommand(flags); err != nil {
				feedback.Errorf("Error during firmware encoding: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	uploadCommand.Flags().StringVarP(&flags.FQBN, "fqbn", "b", "", "Device fqbn")
	uploadCommand.Flags().StringVarP(&flags.file, "file", "", "", "Binary file (.bin) to be encoded")
	uploadCommand.MarkFlagRequired("fqbn")
	uploadCommand.MarkFlagRequired("file")
	return uploadCommand
}

func runEncodeCommand(flags *encodeBinaryFlags) error {
	params := &ota.EncodeParams{
		FQBN: flags.FQBN,
		File: flags.file,
	}
	otafile, err := ota.Encode(params)
	if err != nil {
		return err
	}

	feedback.Print(fmt.Sprintf("Encode successfully performed. File: %s", *otafile))

	return nil
}
