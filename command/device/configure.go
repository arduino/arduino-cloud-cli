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
	"time"

	"github.com/arduino/arduino-cloud-cli/arduino/cli"
	configurationprotocol "github.com/arduino/arduino-cloud-cli/internal/board-protocols/configuration-protocol"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/configuration-protocol/cborcoders"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/transport"
	"github.com/arduino/arduino-cloud-cli/internal/serial"
	"github.com/sirupsen/logrus"
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

	nc := NewNetworkConfigure(extInterface)
	err = nc.Run(ctx, NetConfig)

	return err
}

type ConfigStatus int

const (
	NoneState ConfigStatus = iota
	WaitForConnection
	WaitingForInitialStatus
	WaitingForNetworkOptions
	ConfigureNetwork
	SendConnectionRequest
	WaitingForConnectionCommandResult
	WaitingForNetworkConfigResult
	End
)

type NetworkConfigure struct {
	state          ConfigStatus
	extInterface   transport.TransportInterface
	configProtocol *configurationprotocol.NetworkConfigurationProtocol
}

func NewNetworkConfigure(extInterface transport.TransportInterface) *NetworkConfigure {
	return &NetworkConfigure{
		extInterface:   extInterface,
		configProtocol: configurationprotocol.NewNetworkConfigurationProtocol(&extInterface),
	}
}

func (nc *NetworkConfigure) Run(ctx context.Context, netConfig *NetConfig) error {
	nc.state = WaitForConnection
	var err error
	for nc.state != End {

		switch nc.state {
		case WaitForConnection:
			err = nc.waitForConnection()
			if err != nil {
				nc.state = End
			}
		case WaitingForInitialStatus:
			err = nc.waitingForInitialStatus()
			if err != nil {
				nc.state = End
			}
		case WaitingForNetworkOptions:
			err = nc.waitingForNetworkOptions()
			if err != nil {
				nc.state = End
			}
		case ConfigureNetwork:
			err = nc.configureNetwork(ctx, netConfig)
			if err != nil {
				nc.state = End
			}
		case SendConnectionRequest:
			err = nc.sendConnectionRequest()
			if err != nil {
				nc.state = End
			}
		case WaitingForConnectionCommandResult:
			err = nc.waitingForConnectionCommandResult()
			if err != nil {
				nc.state = End
			}
		case WaitingForNetworkConfigResult:
			err = nc.waitingForNetworkConfigResult()
			if err != nil {
				nc.state = End
			}
		}

	}

	nc.configProtocol.Close()
	return err
}

func (nc *NetworkConfigure) waitForConnection() error {
	if nc.extInterface.Connected() {
		nc.state = WaitingForInitialStatus
	}
	return nil
}

func (nc *NetworkConfigure) waitingForInitialStatus() error {
	logrus.Info("NetworkConfigure: waiting for initial status from device")
	res, err := nc.configProtocol.ReceiveData(30)
	if err != nil {
		return fmt.Errorf("communication error: %w, please check the NetworkConfigurator lib is activated in the sketch", err)
	}

	if res == nil {
		nc.state = WaitingForNetworkOptions
	} else if res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		if status.Status == 1 {
			nc.state = WaitingForInitialStatus
		} else if status.Status == -6 || status.Status <= -101 {
			newState, err := nc.handleStatusMessage(status.Status)
			if err != nil {
				return err
			}
			if newState != NoneState {
				nc.state = newState
			}
		} else {
			nc.state = WaitingForNetworkOptions
		}

	} else if res.Type() == cborcoders.WiFiNetworksType {
		nc.state = ConfigureNetwork
	}

	return nil
}

func (nc *NetworkConfigure) waitingForNetworkOptions() error {
	logrus.Info("NetworkConfigure: waiting for network options from device")
	res, err := nc.configProtocol.ReceiveData(30)
	if err != nil {
		return err
	}

	if res != nil {
		if res.Type() == cborcoders.WiFiNetworksType {
			nc.state = ConfigureNetwork
		} else if res.Type() == cborcoders.ProvisioningStatusMessageType {
			status := res.ToProvisioningStatusMessage()
			if status.Status == 1 {
				nc.state = WaitingForInitialStatus
			} else {
				newState, err := nc.handleStatusMessage(status.Status)
				if err != nil {
					return err
				}
				if newState != NoneState {
					nc.state = newState
				}
			}
		}
	}

	return nil
}

func (nc *NetworkConfigure) configureNetwork(ctx context.Context, c *NetConfig) error {
	var cmd cborcoders.Cmd
	if c.Type == 1 { // WiFi
		cmd = cborcoders.From(cborcoders.ProvisioningWifiConfigMessage{
			SSID: c.WiFi.SSID,
			PWD:  c.WiFi.PWD,
		})
	} else if c.Type == 2 { // Ethernet
		cmd = cborcoders.From(cborcoders.ProvisioningEthernetConfigMessage{
			Static_ip:       c.Eth.IP.Bytes[:],
			Dns:             c.Eth.DNS.Bytes[:],
			Gateway:         c.Eth.Gateway.Bytes[:],
			Netmask:         c.Eth.Netmask.Bytes[:],
			Timeout:         c.Eth.Timeout,
			ResponseTimeout: c.Eth.ResponseTimeout,
		})
	} else if c.Type == 3 { // NB-IoT
		cmd = cborcoders.From(cborcoders.ProvisioningNBConfigMessage{
			PIN:   c.NB.PIN,
			Apn:   c.NB.APN,
			Login: c.NB.Login,
			Pass:  c.NB.Pass,
		})
	} else if c.Type == 4 { // GSM
		cmd = cborcoders.From(cborcoders.ProvisioningGSMConfigMessage{
			PIN:   c.GSM.PIN,
			Apn:   c.GSM.APN,
			Login: c.GSM.Login,
			Pass:  c.GSM.Pass,
		})
	} else if c.Type == 5 { // LoRa
		cmd = cborcoders.From(cborcoders.ProvisioningLoRaConfigMessage{
			AppEui:      c.Lora.AppEUI,
			AppKey:      c.Lora.AppKey,
			Band:        c.Lora.Band,
			ChannelMask: c.Lora.ChannelMask,
			DeviceClass: c.Lora.DeviceClass,
		})
	} else if c.Type == 6 { // CAT-M1
		cmd = cborcoders.From(cborcoders.ProvisioningCATM1ConfigMessage{
			PIN:   c.CATM1.PIN,
			Apn:   c.CATM1.APN,
			Login: c.CATM1.Login,
			Pass:  c.CATM1.Pass,
			Band:  c.CATM1.Band,
		})
	} else if c.Type == 7 { // Cellular
		cmd = cborcoders.From(cborcoders.ProvisioningCellularConfigMessage{
			PIN:   c.CellularSetting.PIN,
			Apn:   c.CellularSetting.APN,
			Login: c.CellularSetting.Login,
			Pass:  c.CellularSetting.Pass,
		})
	} else {
		return errors.New("invalid configuration type")
	}

	err := nc.configProtocol.SendData(cmd)
	if err != nil {
		return err
	}

	nc.state = SendConnectionRequest
	sleepCtx(ctx, 1*time.Second)
	return nil
}

func (nc *NetworkConfigure) sendConnectionRequest() error {
	connectMessage := cborcoders.From(cborcoders.ProvisioningCommandsMessage{Command: configurationprotocol.Commands["Connect"]})
	err := nc.configProtocol.SendData(connectMessage)
	if err != nil {
		return err
	}
	nc.state = WaitingForConnectionCommandResult
	return nil
}

func (nc *NetworkConfigure) waitingForConnectionCommandResult() error {
	res, err := nc.configProtocol.ReceiveData(60)
	if err != nil {
		return err
	}

	if res != nil && res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		if status.Status == 1 {
			nc.state = WaitingForNetworkConfigResult
		} else {
			newState, err := nc.handleStatusMessage(status.Status)
			if err != nil {
				return err
			}
			if newState != NoneState {
				nc.state = newState
			}
		}
	}

	return nil
}

func (nc *NetworkConfigure) waitingForNetworkConfigResult() error {
	res, err := nc.configProtocol.ReceiveData(200)
	if err != nil {
		return err
	}

	if res != nil && res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()

		if status.Status == 2 {
			nc.state = End
		} else {
			newState, err := nc.handleStatusMessage(status.Status)
			if err != nil {
				return err
			}
			if newState != NoneState {
				nc.state = newState
			}
		}
	}

	return nil
}

func (nc *NetworkConfigure) printNetworkOption(msg *cborcoders.Cmd) {
	if msg.Type() == cborcoders.WiFiNetworksType {
		networks := msg.ToWiFiNetworks()
		for _, network := range networks {
			fmt.Printf("SSID: %s, RSSI %d \n", network.SSID, network.RSSI)
		}
	}
}

func (nc *NetworkConfigure) handleStatusMessage(status int16) (ConfigStatus, error) {
	statusMessage := configurationprotocol.StatusBoard[status]
	logrus.Debugf("NetworkConfigure: status message received: %s", statusMessage)

	switch statusMessage {
	case "Connecting":
		return NoneState, nil
	case "Connected":
		return NoneState, nil
	case "Resetted":
		return NoneState, nil
	case "Scanning for WiFi networks":
		return WaitingForNetworkOptions, nil
	case "Failed to connect":
		return NoneState, errors.New("connection failed invalid credentials or network configuration")
	case "Disconnected":
		return NoneState, nil
	case "Parameters not provided":
		return ConfigureNetwork, nil
	case "Invalid parameters":
		return NoneState, errors.New("the provided parameters for network configuration are invalid")
	case "Cannot execute anew request while another is pending":
		return NoneState, errors.New("board is busy, restart the board and try again")
	case "Invalid request":
		return NoneState, errors.New("invalid request sent to the board")
	case "Internet not available":
		return NoneState, errors.New("internet not available, check your network connection")
	case "HW Error connectivity module":
		return NoneState, errors.New("hardware error in connectivity module, check the board")
	case "HW Connectivity Module stopped":
		return NoneState, errors.New("hardware connectivity module stopped, restart the board and check your sketch")
	case "Error initializing secure element":
		return NoneState, errors.New("error initializing secure element, check the board and try again")
	case "Error configuring secure element":
		return NoneState, errors.New("error configuring secure element, check the board and try again")
	case "Error locking secure element":
		return NoneState, errors.New("error locking secure element, check the board and try again")
	case "Error generating UHWID":
		return NoneState, errors.New("error generating UHWID, check the board and try again")
	case "Error storage begin module":
		return NoneState, errors.New("error beginning storage module, check the board storage partitioning and try again")
	case "Fail to partition the storage":
		return NoneState, errors.New("failed to partition the storage, check the board storage and try again")
	case "Generic error":
		return NoneState, errors.New("generic error, check the board and try again")
	default:
		return NoneState, errors.New("generic error, check the board and try again")
	}

}
