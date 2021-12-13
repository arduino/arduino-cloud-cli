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

var createFlags struct {
	port string
	name string
	fqbn string
}

func initCreateCommand() *cobra.Command {
	createCommand := &cobra.Command{
		Use:   "create",
		Short: "Create a device",
		Long:  "Create a device for Arduino IoT Cloud",
		Run:   runCreateCommand,
	}
	createCommand.Flags().StringVarP(&createFlags.port, "port", "p", "", "Device port")
	createCommand.Flags().StringVarP(&createFlags.name, "name", "n", "", "Device name")
	createCommand.Flags().StringVarP(&createFlags.fqbn, "fqbn", "b", "", "Device fqbn")
	createCommand.MarkFlagRequired("name")
	return createCommand
}

func runCreateCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Creating device with name %s", createFlags.name)

	params := &device.CreateParams{
		Name: createFlags.name,
	}
	if createFlags.port != "" {
		params.Port = &createFlags.port
	}
	if createFlags.fqbn != "" {
		params.FQBN = &createFlags.fqbn
	}

	dev, err := device.Create(params)
	if err != nil {
		feedback.Errorf("Error during device create: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.PrintResult(createResult{dev})
}

type createResult struct {
	device *device.DeviceInfo
}

func (r createResult) Data() interface{} {
	return r.device
}

func (r createResult) String() string {
	return fmt.Sprintf(
		"name: %s\nid: %s\nboard: %s\nserial-number: %s\nfqbn: %s",
		r.device.Name,
		r.device.ID,
		r.device.Board,
		r.device.Serial,
		r.device.FQBN,
	)
}
