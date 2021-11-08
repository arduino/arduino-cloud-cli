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

package device

import (
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/device"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var deleteFlags struct {
	id   string
	tags map[string]string
}

func initDeleteCommand() *cobra.Command {
	deleteCommand := &cobra.Command{
		Use:   "delete",
		Short: "Delete a device",
		Long:  "Delete a device from Arduino IoT Cloud",
		Run:   runDeleteCommand,
	}
	deleteCommand.Flags().StringVarP(&deleteFlags.id, "id", "i", "", "Device ID")
	deleteCommand.Flags().StringToStringVar(
		&deleteFlags.tags,
		"tags",
		nil,
		"Comma-separated list of tags with format <key>=<value>.\n"+
			"Delete all devices that match the provided tags.\n"+
			"Mutually exclusive with `--id`.",
	)
	return deleteCommand
}

func runDeleteCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Deleting device %s\n", deleteFlags.id)

	params := &device.DeleteParams{Tags: deleteFlags.tags}
	if deleteFlags.id != "" {
		params.ID = &deleteFlags.id
	}

	err := device.Delete(params)
	if err != nil {
		feedback.Errorf("Error during device delete: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Info("Device successfully deleted")
}
