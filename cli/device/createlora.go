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

var createLoraFlags struct {
	port          string
	name          string
	fqbn          string
	frequencyPlan string
}

func initCreateLoraCommand() *cobra.Command {
	createLoraCommand := &cobra.Command{
		Use:   "create-lora",
		Short: "Create a LoRa device",
		Long:  "Create a LoRa device for Arduino IoT Cloud",
		Run:   runCreateLoraCommand,
	}
	createLoraCommand.Flags().StringVarP(&createLoraFlags.port, "port", "p", "", "Device port")
	createLoraCommand.Flags().StringVarP(&createLoraFlags.name, "name", "n", "", "Device name")
	createLoraCommand.Flags().StringVarP(&createLoraFlags.fqbn, "fqbn", "b", "", "Device fqbn")
	createLoraCommand.Flags().StringVarP(&createLoraFlags.frequencyPlan, "frequency-plan", "f", "",
		"ID of the LoRa frequency plan to use. Run the 'device list-frequency-plans' command to obtain a list of valid plans.")
	createLoraCommand.MarkFlagRequired("name")
	createLoraCommand.MarkFlagRequired("frequency-plan")
	return createLoraCommand
}

func runCreateLoraCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Creating LoRa device with name %s", createLoraFlags.name)

	params := &device.CreateLoraParams{
		CreateParams: device.CreateParams{
			Name: createLoraFlags.name,
		},
		FrequencyPlan: createLoraFlags.frequencyPlan,
	}
	if createLoraFlags.port != "" {
		params.Port = &createLoraFlags.port
	}
	if createLoraFlags.fqbn != "" {
		params.FQBN = &createLoraFlags.fqbn
	}

	dev, err := device.CreateLora(params)
	if err != nil {
		feedback.Errorf("Error during device create-lora: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.PrintResult(createLoraResult{dev})
}

type createLoraResult struct {
	device *device.DeviceLoraInfo
}

func (r createLoraResult) Data() interface{} {
	return r.device
}

func (r createLoraResult) String() string {
	return fmt.Sprintf(
		"name: %s\nid: %s\nboard: %s\nserial-number: %s\nfqbn: %s"+
			"\napp-eui: %s\napp-key: %s\neui: %s",
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
