// This file is part of arduino-cloud-cli.
//
// Copyright (C) ARDUINO SRL (http://www.arduino.cc)
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
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/arduino-cloud-cli/command/device"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type showFlags struct {
	deviceId string
}

func initShowCommand() *cobra.Command {
	flags := &showFlags{}
	showCommand := &cobra.Command{
		Use:   "show",
		Short: "Show device properties",
		Long:  "Show device properties on Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runShowCommand(flags); err != nil {
				feedback.Errorf("Error during device show: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	showCommand.Flags().StringVarP(&flags.deviceId, "device-id", "d", "", "device ID")

	showCommand.MarkFlagRequired("device-id")

	return showCommand
}

func runShowCommand(flags *showFlags) error {
	logrus.Info("Show device")

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	dev, _, err := device.Show(context.TODO(), flags.deviceId, cred)
	if err != nil {
		return err
	}

	feedback.PrintResult(showResult{dev})
	return nil
}

type showResult struct {
	device *device.DeviceInfo
}

func (r showResult) Data() interface{} {
	return r.device
}

func (r showResult) String() string {
	if r.device == nil {
		return "No device found."
	}
	t := table.New()
	t.SetHeader("Name", "ID", "Board", "FQBN", "SerialNumber", "Status", "Connection type", "Thing", "Tags")
	t.AddRow(
		r.device.Name,
		r.device.ID,
		r.device.Board,
		r.device.FQBN,
		r.device.Serial,
		dereferenceString(r.device.Status),
		dereferenceString(r.device.ConnectionType),
		dereferenceString(r.device.ThingID),
		strings.Join(r.device.Tags, ","),
	)
	return t.Render()
}
