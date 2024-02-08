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

type massCreateFlags struct {
	name  string
	fqbn  string
}

func initMassCreateCommand() *cobra.Command {
	flags := &massCreateFlags{}
	createCommand := &cobra.Command{
		Use:   "mass-create",
		Short: "Mass create a set of devices provisioning the onboard secure element with a valid certificate",
		Long:  "Mass create a set of devices for Arduino IoT Cloud provisioning the onboard secure element with a valid certificate",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runMassCreateCommand(flags); err != nil {
				feedback.Errorf("Error during device create: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	createCommand.Flags().StringVarP(&flags.name, "name", "n", "", "Base device name")
	createCommand.Flags().StringVarP(&flags.fqbn, "fqbn", "b", "", "Device fqbn")
	createCommand.MarkFlagRequired("name")
	return createCommand
}


func runMassCreateCommand(flags *massCreateFlags) error {
	logrus.Infof("Mass provisioning devices. Base name: %s", flags.name)

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	ctx, cancel := cleanup.InterruptableContext(context.Background())
	defer cancel()

	boards, err := device.ListAllConnectedBoardsWithCrypto(&flags.fqbn)
	if err != nil {
		return err
	}
	if len(boards) == 0 {
		return fmt.Errorf("no boards of type %s detected", flags.fqbn)
	}

	var results []*device.DeviceInfo
	for idx, board := range boards {
		if len(board.Address) == 0{
			continue
		}
		logrus.Infof("Provisioning board on port: %s", board.Address)
		bname := fmt.Sprintf("%s-%d", flags.name, idx)
		params := &device.CreateParams{
			Name: bname,
			Port: &board.Address,
			FQBN: &board.Fqbn,
		}
	
		dev, err := device.Create(ctx, params, cred)
		if err != nil {
			return err
		}

		results = append(results, dev)
	}

	feedback.PrintResult(massCreateResult{results})

	return nil
}

type massCreateResult struct {
	devices []*device.DeviceInfo
}

func (r massCreateResult) Data() interface{} {
	return r.devices
}

func (r massCreateResult) String() string {
	var result string
	for _, device := range r.devices {
		result += fmt.Sprintf(
			"name: %s\nid: %s\nboard: %s\nserial_number: %s\nfqbn: %s\n-------------\n",
			device.Name,
			device.ID,
			device.Board,
			device.Serial,
			device.FQBN,
		)
	}
	return result
}
