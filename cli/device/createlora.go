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

type createLoraFlags struct {
	port          string
	name          string
	fqbn          string
	frequencyPlan string
}

func initCreateLoraCommand() *cobra.Command {
	flags := &createLoraFlags{}
	createLoraCommand := &cobra.Command{
		Use:   "create-lora",
		Short: "Create a LoRa device",
		Long:  "Create a LoRa device for Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runCreateLoraCommand(flags); err != nil {
				feedback.Errorf("Error during device create-lora: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	createLoraCommand.Flags().StringVarP(&flags.port, "port", "p", "", "Device port")
	createLoraCommand.Flags().StringVarP(&flags.name, "name", "n", "", "Device name")
	createLoraCommand.Flags().StringVarP(&flags.fqbn, "fqbn", "b", "", "Device fqbn")
	createLoraCommand.Flags().StringVarP(&flags.frequencyPlan, "frequency-plan", "f", "",
		"ID of the LoRa frequency plan to use. Run the 'device list-frequency-plans' command to obtain a list of valid plans.")
	createLoraCommand.MarkFlagRequired("name")
	createLoraCommand.MarkFlagRequired("frequency-plan")
	return createLoraCommand
}

func runCreateLoraCommand(flags *createLoraFlags) error {
	logrus.Infof("Creating LoRa device with name %s", flags.name)

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	params := &device.CreateLoraParams{
		CreateParams: device.CreateParams{
			Name: flags.name,
		},
		FrequencyPlan: flags.frequencyPlan,
	}
	if flags.port != "" {
		params.Port = &flags.port
	}
	if flags.fqbn != "" {
		params.FQBN = &flags.fqbn
	}

	ctx, cancel := cleanup.InterruptableContext(context.Background())
	defer cancel()

	dev, err := device.CreateLora(ctx, params, cred)
	if err != nil {
		return err
	}

	feedback.PrintResult(createLoraResult{dev})
	return nil
}

type createLoraResult struct {
	device *device.DeviceLoraInfo
}

func (r createLoraResult) Data() interface{} {
	return r.device
}

func (r createLoraResult) String() string {
	return fmt.Sprintf(
		"name: %s\nid: %s\nboard: %s\nserial_number: %s\nfqbn: %s"+
			"\napp_eui: %s\napp_key: %s\neui: %s",
		r.device.Name,
		r.device.ID,
		r.device.Board,
		r.device.Serial,
		r.device.FQBN,
		r.device.AppEUI,
		r.device.AppKey,
		r.device.EUI,
	)
}
