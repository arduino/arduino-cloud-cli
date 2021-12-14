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
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/device"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var createGenericFlags struct {
	name string
	fqbn string
}

func initCreateGenericCommand() *cobra.Command {
	createGenericCommand := &cobra.Command{
		Use:   "create-generic",
		Short: "Create a generic device",
		Long:  "Create a generic device for Arduino IoT Cloud",
		Run:   runCreateGenericCommand,
	}
	createGenericCommand.Flags().StringVarP(&createGenericFlags.name, "name", "n", "", "Device name")
	createGenericCommand.Flags().StringVarP(&createGenericFlags.fqbn, "fqbn", "b", "generic:generic:generic", "Device fqbn")
	createGenericCommand.MarkFlagRequired("name")
	return createGenericCommand
}

func runCreateGenericCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Creating generic device with name %s", createGenericFlags.name)

	params := &device.CreateGenericParams{
		Name: createGenericFlags.name,
		FQBN: createGenericFlags.fqbn,
	}

	dev, err := device.CreateGeneric(params)
	if err != nil {
		feedback.Errorf("Error during device create-generic: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.PrintResult(createGenericResult{dev})
}

type createGenericResult struct {
	device *device.DeviceGenericInfo
}

func (r createGenericResult) Data() interface{} {
	return r.device
}

func (r createGenericResult) String() string {
	return fmt.Sprintf(
		"id: %s\nsecret_key: %s\nname: %s\nboard: %s\nserial_number: %s\nfqbn: %s",
		r.device.ID,
		r.device.Password,
		r.device.Name,
		r.device.Board,
		r.device.Serial,
		r.device.FQBN,
	)
}
