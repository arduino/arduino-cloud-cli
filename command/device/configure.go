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
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/arduino/arduino-cloud-cli/arduino/cli"
	configurationprotocol "github.com/arduino/arduino-cloud-cli/internal/board-protocols/configuration-protocol"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/transport"
	"github.com/arduino/arduino-cloud-cli/internal/serial"
)

func NetConfigure(ctx context.Context, boardFilters *CreateParams, NetConfig *NetConfig) error {
	comm, err := cli.NewCommander()
	if err != nil {
		return err
	}

	ports, err := comm.BoardList(ctx)
	if err != nil {
		return err
	}

	board := boardFromPorts(ports, boardFilters)
	if board == nil {
		err = errors.New("no board found")
		return err
	}
	var extInterface transport.TransportInterface
	extInterface = &serial.Serial{}
	configProtocol := configurationprotocol.NewNetworkConfigurationProtocol(&extInterface)

	err = configProtocol.Connect(board.address)
	if err != nil {
		return err
	}

	nc := NewNetworkConfigure(extInterface, configProtocol)
	err = nc.Run(ctx, NetConfig)

	return err
}

func GetInputFromMenu(config *NetConfig) error {

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

func getWiFiSetting() WiFiSetting {
	var wifi WiFiSetting
	fmt.Print("Enter SSID: ")
	fmt.Scanln(&wifi.SSID)
	fmt.Print("Enter Password: ")
	fmt.Scanln(&wifi.PWD)
	return wifi
}

func getEthernetSetting() EthernetSetting {
	var eth EthernetSetting
	fmt.Println("Do you want to use DHCP? (yes/no): ")
	var useDHCP string
	fmt.Scanln(&useDHCP)
	if useDHCP == "yes" || useDHCP == "y" {
		eth.IP = IPAddr{Type: 0, Bytes: [16]byte{}}
		eth.Gateway = IPAddr{Type: 0, Bytes: [16]byte{}}
		eth.Netmask = IPAddr{Type: 0, Bytes: [16]byte{}}
		eth.DNS = IPAddr{Type: 0, Bytes: [16]byte{}}
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

func getIPAddr() IPAddr {
	var ip IPAddr
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

func getCellularSetting() CellularSetting {
	var cellular CellularSetting
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

func getCatM1Setting() CATM1Setting {
	var catm1 CATM1Setting
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

func getLoraSetting() LoraSetting {
	var lora LoraSetting
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

type NetworkConfigure struct {
	configStates   *ConfigurationStates
	configProtocol *configurationprotocol.NetworkConfigurationProtocol
}

func NewNetworkConfigure(extInterface transport.TransportInterface, configProtocol *configurationprotocol.NetworkConfigurationProtocol) *NetworkConfigure {
	return &NetworkConfigure{
		configStates:   NewConfigurationStates(extInterface, configProtocol),
		configProtocol: configProtocol,
	}
}

func (nc *NetworkConfigure) Run(ctx context.Context, netConfig *NetConfig) error {
	state := WaitForConnection
	nextState := state
	var err error

	for state != End {

		switch state {
		case WaitForConnection:
			nextState, err = nc.configStates.WaitForConnection()
			if err != nil {
				nextState = End
			}
		case WaitingForInitialStatus:
			nextState, err = nc.configStates.WaitingForInitialStatus()
			if err != nil {
				nextState = End
			}
		case WaitingForNetworkOptions:
			nextState, err = nc.configStates.WaitingForNetworkOptions()
			if err != nil {
				nextState = End
			}
		case BoardReady:
			nextState = ConfigureNetwork
		case ConfigureNetwork:
			nextState, err = nc.configStates.ConfigureNetwork(ctx, netConfig)
			if err != nil {
				nextState = End
			}
		case SendConnectionRequest:
			nextState, err = nc.configStates.SendConnectionRequest()
			if err != nil {
				nextState = End
			}
		case WaitingForConnectionCommandResult:
			nextState, err = nc.configStates.WaitingForConnectionCommandResult()
			if err != nil {
				nextState = End
			}
		case MissingParameter:
			nextState = ConfigureNetwork
		case WaitingForNetworkConfigResult:
			nextState, err = nc.configStates.WaitingForNetworkConfigResult()
			if err != nil {
				nextState = End
			}
		}

		state = nextState

	}

	nc.configProtocol.Close()
	return err
}
