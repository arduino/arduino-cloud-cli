// This file is part of arduino-cloud-cli.
//
// Copyright (C) 2025 ARDUINO SA (http://www.arduino.cc/)
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

type migrateFlags struct {
	port string
	fqbn string
}

func initMigrateCommand() *cobra.Command {
	flags := &migrateFlags{}
	createCommand := &cobra.Command{
		Use:   "migrate",
		Short: "Set-up the device to enable Bluetooth provisioning",
		Long:  "Set-up the device to enable Bluetooth provisioning",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runMigrateCommand(flags); err != nil {
				feedback.Errorf("Error during device configuration: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	createCommand.Flags().StringVarP(&flags.port, "port", "p", "", "Device port")
	createCommand.Flags().StringVarP(&flags.fqbn, "fqbn", "b", "", "Device fqbn")
	return createCommand
}

func runMigrateCommand(flags *migrateFlags) error {
	logrus.Infof("Setting up the device to enable Bluetooth provisioning")

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	boardFilterParams := &device.MigrateParams{}

	if flags.port != "" {
		boardFilterParams.Port = &flags.port
	}
	if flags.fqbn != "" {
		boardFilterParams.FQBN = &flags.fqbn
	}

	ctx, cancel := cleanup.InterruptableContext(context.Background())
	defer cancel()
	feedback.Print("Starting device setup for Bluetooth provisioning...")
	err = device.Migrate(ctx, boardFilterParams, cred)
	if err != nil {
		return err
	}
	feedback.Print("Device setup for Bluetooth provisioning successfully completed.")
	return nil
}
