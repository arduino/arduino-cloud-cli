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
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

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
		getInputFromMenu(netParams)
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

func getInputFromMenu(config *device.NetConfig) error {

	switch config.Type {
	case 1:
		config.WiFi = getWiFiSetting()
	case 2:
		config.Eth = getEthernetSetting()
	case 3:
		config.NB = getCellularSetting()
	case 4:
		config.GSM = getCellularSetting()
	case 5:
		config.Lora = getLoraSetting()
	case 6:
		config.CATM1 = getCatM1Setting()
	case 7:
		config.CellularSetting = getCellularSetting()
	default:
		return errors.New("invalid connection type, please try again")
	}
	return nil
}

func getWiFiSetting() device.WiFiSetting {
	var wifi device.WiFiSetting
	fmt.Print("Enter SSID: ")
	fmt.Scanln(&wifi.SSID)
	fmt.Print("Enter Password: ")
	fmt.Scanln(&wifi.PWD)
	return wifi
}

func getEthernetSetting() device.EthernetSetting {
	var eth device.EthernetSetting
	fmt.Println("Do you want to use DHCP? (yes/no): ")
	var useDHCP string
	fmt.Scanln(&useDHCP)
	if useDHCP == "yes" || useDHCP == "y" {
		eth.IP = device.IPAddr{Type: 0, Bytes: [16]byte{}}
		eth.Gateway = device.IPAddr{Type: 0, Bytes: [16]byte{}}
		eth.Netmask = device.IPAddr{Type: 0, Bytes: [16]byte{}}
		eth.DNS = device.IPAddr{Type: 0, Bytes: [16]byte{}}
	} else {
		fmt.Println("Enter IP Address: ")
		eth.IP = getIPAddr()
		fmt.Println("Enter DNS: ")
		eth.DNS = getIPAddr()
		fmt.Println("Enter Gateway: ")
		eth.Gateway = getIPAddr()
		fmt.Println("Enter Netmask: ")
		eth.Netmask = getIPAddr()
	}

	return eth
}

func getIPAddr() device.IPAddr {
	var ip device.IPAddr
	var ipString string
	fmt.Scanln(&ipString)
	if ipString == "" {
		return ip
	}
	if strings.Count(ipString, ":") > 0 {
		ip.Type = 1 // IPv6
	} else {
		ip.Type = 0 // IPv4
	}
	ip.Bytes = [16]byte(net.ParseIP(ipString).To16())
	return ip
}

func getCellularSetting() device.CellularSetting {
	var cellular device.CellularSetting
	fmt.Println("Enter PIN: ")
	fmt.Scanln(&cellular.PIN)
	fmt.Print("Enter APN: ")
	fmt.Scanln(&cellular.APN)
	fmt.Print("Enter Login: ")
	fmt.Scanln(&cellular.Login)
	fmt.Print("Enter Password: ")
	fmt.Scanln(&cellular.Pass)
	return cellular
}

func getCatM1Setting() device.CATM1Setting {
	var catm1 device.CATM1Setting
	fmt.Print("Enter PIN: ")
	fmt.Scanln(&catm1.PIN)
	fmt.Print("Enter APN: ")
	fmt.Scanln(&catm1.APN)
	fmt.Print("Enter Login: ")
	fmt.Scanln(&catm1.Login)
	fmt.Print("Enter Password: ")
	fmt.Scanln(&catm1.Pass)
	return catm1
}

func getLoraSetting() device.LoraSetting {
	var lora device.LoraSetting
	fmt.Print("Enter AppEUI: ")
	fmt.Scanln(&lora.AppEUI)
	fmt.Print("Enter AppKey: ")
	fmt.Scanln(&lora.AppKey)
	fmt.Print("Enter Band (Byte hex format): ")
	fmt.Scanln(&lora.Band)
	fmt.Print("Enter Channel Mask: ")
	fmt.Scanln(&lora.ChannelMask)
	fmt.Print("Enter Device Class: ")
	fmt.Scanln(&lora.DeviceClass)
	return lora
}
