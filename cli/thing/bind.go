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

package thing

import (
	"context"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/thing"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type bindFlags struct {
	id       string
	deviceID string
}

func initBindCommand() *cobra.Command {
	flags := &bindFlags{}
	bindCommand := &cobra.Command{
		Use:   "bind",
		Short: "Bind a thing to a device",
		Long:  "Bind a thing to a device on Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runBindCommand(flags); err != nil {
				feedback.Errorf("Error during thing bind: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	bindCommand.Flags().StringVarP(&flags.id, "id", "i", "", "Thing ID")
	bindCommand.Flags().StringVarP(&flags.deviceID, "device-id", "d", "", "Device ID")
	bindCommand.MarkFlagRequired("id")
	bindCommand.MarkFlagRequired("device-id")
	return bindCommand
}

func runBindCommand(flags *bindFlags) error {
	logrus.Infof("Binding thing %s to device %s", flags.id, flags.deviceID)

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	params := &thing.BindParams{
		ID:       flags.id,
		DeviceID: flags.deviceID,
	}
	if err = thing.Bind(context.TODO(), params, cred); err != nil {
		return err
	}

	logrus.Info("Thing-Device bound successfully updated")
	return nil
}
