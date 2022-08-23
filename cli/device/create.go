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
	"context"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/device"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.bug.st/cleanup"
)

type createFlags struct {
	port string
	name string
	fqbn string
}

func initCreateCommand() *cobra.Command {
	flags := &createFlags{}
	createCommand := &cobra.Command{
		Use:   "create",
		Short: "Create a device provisioning the onboard secure element with a valid certificate",
		Long:  "Create a device for Arduino IoT Cloud provisioning the onboard secure element with a valid certificate",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runCreateCommand(flags); err != nil {
				feedback.Errorf("Error during device create: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	createCommand.Flags().StringVarP(&flags.port, "port", "p", "", "Device port")
	createCommand.Flags().StringVarP(&flags.name, "name", "n", "", "Device name")
	createCommand.Flags().StringVarP(&flags.fqbn, "fqbn", "b", "", "Device fqbn")
	createCommand.MarkFlagRequired("name")
	return createCommand
}

func runCreateCommand(flags *createFlags) error {
	logrus.Infof("Creating device with name %s", flags.name)

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	params := &device.CreateParams{
		Name: flags.name,
	}
	if flags.port != "" {
		params.Port = &flags.port
	}
	if flags.fqbn != "" {
		params.FQBN = &flags.fqbn
	}

	ctx, cancel := cleanup.InterruptableContext(context.Background())
	defer cancel()

	dev, err := device.Create(ctx, params, cred)
	if err != nil {
		return err
	}

	feedback.PrintResult(createResult{dev})
	return nil
}

type createResult struct {
	device *device.DeviceInfo
}

func (r createResult) Data() interface{} {
	return r.device
}

func (r createResult) String() string {
	return fmt.Sprintf(
		"name: %s\nid: %s\nboard: %s\nserial_number: %s\nfqbn: %s",
		r.device.Name,
		r.device.ID,
		r.device.Board,
		r.device.Serial,
		r.device.FQBN,
	)
}
