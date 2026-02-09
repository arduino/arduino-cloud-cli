// This file is part of arduino-cloud-cli.
//
// Copyright (C) ARDUINO SRL (http://www.arduino.cc/)
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
	"encoding/json"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/device"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.bug.st/cleanup"
)

type netConfigurationFlags struct {
	port           string
	connectionType int32
	fqbn           string
	configFile     string
}

func initConfigureCommand() *cobra.Command {
	flags := &netConfigurationFlags{}
	createCommand := &cobra.Command{
		Use:   "configure",
		Short: "Configure the network settings of a device running a sketch with the Network Configurator lib enabled",
		Long:  "Configure the network settings of a device running a sketch with the Network Configurator lib enabled",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runConfigureCommand(flags); err != nil {
				feedback.Errorf("Error during device configuration: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	createCommand.Flags().StringVarP(&flags.port, "port", "p", "", "Device port")
	createCommand.Flags().StringVarP(&flags.fqbn, "fqbn", "b", "", "Device fqbn")
	createCommand.Flags().Int32VarP(&flags.connectionType, "connection", "c", 0, "Device connection type (1: WiFi, 2: Ethernet, 3: NB-IoT, 4: GSM, 5: LoRaWan, 6:CAT-M1, 7: Cellular)")
	createCommand.Flags().StringVarP(&flags.configFile, "config-file", "f", "", "Path to the configuration file (optional). View online documentation for the format")
	createCommand.MarkFlagRequired("connection")

	return createCommand
}

func runConfigureCommand(flags *netConfigurationFlags) error {
	logrus.Infof("Configuring device with connection type %d", flags.connectionType)

	netParams := &device.NetConfig{
		Type: flags.connectionType,
	}

	if flags.configFile != "" {
		file, err := os.ReadFile(flags.configFile)
		if err != nil {
			logrus.Errorf("Error reading file %s: %v", flags.configFile, err)
			return err
		}
		err = json.Unmarshal(file, &netParams)
		if err != nil {
			logrus.Errorf("Error parsing JSON from file %s: %v", flags.configFile, err)
			return err
		}
	} else {
		feedback.Print("Insert network configuration")
		device.GetInputFromMenu(netParams)
	}

	boardFilterParams := &device.CreateParams{}

	if flags.port != "" {
		boardFilterParams.Port = &flags.port
	}
	if flags.fqbn != "" {
		boardFilterParams.FQBN = &flags.fqbn
	}

	ctx, cancel := cleanup.InterruptableContext(context.Background())
	defer cancel()
	feedback.Print("Starting network configuration...")
	err := device.NetConfigure(ctx, boardFilterParams, netParams)
	if err != nil {
		return err
	}
	feedback.Print("Network configuration successfully completed.")
	return nil
}
