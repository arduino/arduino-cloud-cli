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
	"os/signal"
	"syscall"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/device"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type createGenericFlags struct {
	name string
	fqbn string
}

func initCreateGenericCommand() *cobra.Command {
	flags := &createGenericFlags{}
	createGenericCommand := &cobra.Command{
		Use:   "create-generic",
		Short: "Create a generic device with password authentication - without secure element - WARNING: less secure",
		Long:  "Create a generic device with password authentication for Arduino IoT Cloud - without secure element - WARNING: less secure",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runCreateGenericCommand(flags); err != nil {
				feedback.Errorf("Error during device create-generic: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	createGenericCommand.Flags().StringVarP(&flags.name, "name", "n", "", "Device name")
	createGenericCommand.Flags().StringVarP(&flags.fqbn, "fqbn", "b", "generic:generic:generic", "Device fqbn")
	createGenericCommand.MarkFlagRequired("name")
	return createGenericCommand
}

func runCreateGenericCommand(flags *createGenericFlags) error {
	logrus.Infof("Creating generic device with name %s", flags.name)

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	params := &device.CreateGenericParams{
		Name: flags.name,
		FQBN: flags.fqbn,
	}

	ctx, canc := context.WithCancel(context.Background())
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		canc()
	}()

	dev, err := device.CreateGeneric(ctx, params, cred)
	if err != nil {
		return err
	}

	feedback.PrintResult(createGenericResult{dev})
	return nil
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
