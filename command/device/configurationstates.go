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
	"strconv"
	"strings"
	"time"

	configurationprotocol "github.com/arduino/arduino-cloud-cli/internal/board-protocols/configuration-protocol"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/configuration-protocol/cborcoders"
	"github.com/sirupsen/logrus"
)

type ConfigStatus int

// This enum represents the different states of the network configuration process
// of the Arduino Board Configuration Protocol.
const (
	NoneState ConfigStatus = iota
	WaitForConnection
	WaitingForInitialStatus
	WaitingForNetworkOptions
	BoardReady
	FlashProvisioningSketch
	GetSketchVersionRequest
	WaitingSketchVersion
	WiFiFWVersionRequest
	WaitingWiFiFWVersion
	RequestBLEMAC
	WaitBLEMAC
	SendInitialTS
	MissingParameter
	IDRequest
	WaitingID
	WaitingSignature
	WaitingPublicKey
	ClaimDevice
	RegisterDevice
	RequestReset
	WaitResetResponse
	GetNetConfigLibVersionRequest
	WaitingNetConfigLibVersion
	ConfigureNetwork
	SendConnectionRequest
	WaitingForConnectionCommandResult
	WaitingForNetworkConfigResult
	WaitingForProvisioningResult
	UnclaimDevice
	End
	ErrorState
)

const (
	CommandResponseTimeoutLong_s  = 60
	CommandResponseTimeoutShort_s = 30
	ConnectResponseTimeout_s      = 200
)

type ConfigurationStates struct {
	configProtocol *configurationprotocol.NetworkConfigurationProtocol
}

func NewConfigurationStates(configProtocol *configurationprotocol.NetworkConfigurationProtocol) *ConfigurationStates {
	return &ConfigurationStates{
		configProtocol: configProtocol,
	}
}

func (c *ConfigurationStates) WaitForConnection() (ConfigStatus, error) {
	if c.configProtocol.Connected() {
		return WaitingForInitialStatus, nil
	}
	return ErrorState, errors.New("impossible to connect with the device")
}

func (c *ConfigurationStates) WaitingForInitialStatus() (ConfigStatus, error) {
	logrus.Info("NetworkConfigure: waiting for initial status from device")
	res, err := c.configProtocol.ReceiveData(CommandResponseTimeoutShort_s)
	if err != nil {
		return ErrorState, fmt.Errorf("communication error: %w, please check the NetworkConfigurator lib is activated in the sketch", err)
	}

	if res == nil {
		return WaitingForNetworkOptions, nil
	}

	if res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		if status.Status == 1 {
			return WaitingForInitialStatus, nil
		}

		if status.Status == -6 || status.Status < -101 {
			return c.HandleStatusMessage(status.Status)
		}
		return WaitingForNetworkOptions, nil
	}

	if res.Type() == cborcoders.WiFiNetworksType {
		return BoardReady, nil
	}

	return WaitingForNetworkOptions, nil
}

// In this state the cli is waiting for the available network options as specified in the
// Arduino Board Configuration Protocol.
func (c *ConfigurationStates) WaitingForNetworkOptions() (ConfigStatus, error) {
	logrus.Info("NetworkConfigure: waiting for network options from device")
	res, err := c.configProtocol.ReceiveData(CommandResponseTimeoutShort_s)
	if err != nil {
		return ErrorState, err
	}

	if res != nil {
		// At the moment of writing, the only type of message that can be received in this state is the
		// WiFiNetworksType, which contains the available WiFi networks list.
		if res.Type() == cborcoders.WiFiNetworksType {
			return BoardReady, nil
		}

		if res.Type() == cborcoders.ProvisioningStatusMessageType {
			status := res.ToProvisioningStatusMessage()
			if status.Status == 1 {
				return WaitingForInitialStatus, nil
			}

			return c.HandleStatusMessage(status.Status)
		}
	}

	return ErrorState, errors.New("timeout: no network options received from the device, please retry enabling the NetworkCofnigurator lib in the sketch")
}

func (c *ConfigurationStates) GetWiFiFWVersionRequest(ctx context.Context) (ConfigStatus, error) {
	logrus.Info("Provisioning V2: Requesting WiFi FW Version")
	getWiFiFWVersionMessage := cborcoders.From(cborcoders.ProvisioningCommandsMessage{Command: configurationprotocol.Commands["GetWiFiFWVersion"]})
	err := c.configProtocol.SendData(getWiFiFWVersionMessage)
	if err != nil {
		return ErrorState, err
	}
	sleepCtx(ctx, 1*time.Second)
	return WaitingWiFiFWVersion, nil
}

func (c *ConfigurationStates) WaitWiFiFWVersion(minWiFiVersion *string) (ConfigStatus, error) {
	res, err := c.configProtocol.ReceiveData(CommandResponseTimeoutLong_s)
	if err != nil {
		return ErrorState, err
	}

	if res == nil {
		return ErrorState, errors.New("provisioning V2: Requesting WiFi FW Version failed")
	}

	if res.Type() == cborcoders.ProvisioningWiFiFWVersionMessageType {
		wifi_version := res.ToProvisioningWiFiFWVersionMessage().WiFiFWVersion
		logrus.Infof("Received WiFi FW Version: %s", wifi_version)
		if minWiFiVersion != nil &&
			c.CompareVersions(wifi_version, *minWiFiVersion) < 0 {
			return ErrorState, fmt.Errorf("provisioning V2: WiFi FW version %s is lower than required minimum %s. Please update the board firmware using Arduino IDE or Arduino CLI", wifi_version, *minWiFiVersion)
		}

		return RequestBLEMAC, nil
	}

	if res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		return c.HandleStatusMessage(status.Status)
	}

	return ErrorState, errors.New("provisioning V2: WiFi FW version not received")
}

/*
 * This function returns
 * - <0 if version1  < version2
 * - =0 if version1 == version2
 * - >0 if version1  > version2
 */
func (c *ConfigurationStates) CompareVersions(version1, version2 string) int {
	version1Tokens := strings.Split(version1, ".")
	version2Tokens := strings.Split(version2, ".")
	if len(version1Tokens) != len(version2Tokens) {
		return -1
	}
	for i := 0; i < len(version1Tokens) && i < len(version2Tokens); i++ {
		version1Num, _ := strconv.Atoi(version1Tokens[i])
		version2Num, _ := strconv.Atoi(version2Tokens[i])
		if version1Num != version2Num {
			return version1Num - version2Num
		}
	}
	return 0
}

func (cs *ConfigurationStates) ConfigureNetwork(ctx context.Context, c *NetConfig) (ConfigStatus, error) {
	logrus.Info("NetworkConfigure: Sending network configuration")
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
			Band:  nil,
		})
	} else if c.Type == 7 { // Cellular
		cmd = cborcoders.From(cborcoders.ProvisioningCellularConfigMessage{
			PIN:   c.CellularSetting.PIN,
			Apn:   c.CellularSetting.APN,
			Login: c.CellularSetting.Login,
			Pass:  c.CellularSetting.Pass,
		})
	} else {
		return ErrorState, errors.New("invalid configuration type")
	}

	err := cs.configProtocol.SendData(cmd)
	if err != nil {
		return ErrorState, err
	}

	sleepCtx(ctx, 1*time.Second)
	return SendConnectionRequest, nil
}

func (c *ConfigurationStates) SendConnectionRequest() (ConfigStatus, error) {
	logrus.Info("NetworkConfigure: Sending connection request")
	connectMessage := cborcoders.From(cborcoders.ProvisioningCommandsMessage{Command: configurationprotocol.Commands["Connect"]})
	err := c.configProtocol.SendData(connectMessage)
	if err != nil {
		return ErrorState, err
	}
	return WaitingForConnectionCommandResult, nil

}

func (c *ConfigurationStates) WaitingForConnectionCommandResult() (ConfigStatus, error) {
	logrus.Info("NetworkConfigure: Waiting for connection command result")
	res, err := c.configProtocol.ReceiveData(CommandResponseTimeoutLong_s)
	if err != nil {
		return ErrorState, err
	}

	if res != nil && res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		if status.Status == 1 {
			return WaitingForNetworkConfigResult, nil
		}

		if status.Status == -4 {
			return ConfigureNetwork, nil
		}

		return c.HandleStatusMessage(status.Status)

	}

	return ErrorState, errors.New("timeout: no confirmation of connection command received from the device, please retry")
}

func (c *ConfigurationStates) WaitingForNetworkConfigResult() (ConfigStatus, error) {
	logrus.Info("NetworkConfigure: Waiting for network configuration result")
	res, err := c.configProtocol.ReceiveData(ConnectResponseTimeout_s)
	if err != nil {
		return ErrorState, err
	}

	if res != nil && res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()

		if status.Status == 2 {
			return End, nil
		}
		//For boards of type Cellular, CAT-M1, GSM e NB-IOT that
		//returns -3 or -101 when the network configuration is invalid
		if status.Status == -3 || status.Status == -101 {
			return ErrorState, errors.New("connection failed: invalid network configuration")
		}
		return c.HandleStatusMessage(status.Status)

	}

	return ErrorState, errors.New("timeout: no result received from the device for network configuration, please retry")
}

// Keep for reference
/*
func (c *ConfigurationStates) printNetworkOption(msg *cborcoders.Cmd) {
	if msg.Type() == cborcoders.WiFiNetworksType {
		networks := msg.ToWiFiNetworks()
		for _, network := range networks {
			fmt.Printf("SSID: %s, RSSI %d \n", network.SSID, network.RSSI)
		}
	}
}
*/

func (c *ConfigurationStates) HandleStatusMessage(status int16) (ConfigStatus, error) {
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
		return ErrorState, errors.New("connection failed: invalid network configuration")
	case "Disconnected":
		return NoneState, nil
	case "Parameters not provided":
		return MissingParameter, nil
	case "Invalid parameters":
		return ErrorState, errors.New("the provided parameters for network configuration are invalid")
	case "Cannot execute anew request while another is pending":
		return ErrorState, errors.New("board is busy, restart the board and try again")
	case "Invalid request":
		return ErrorState, errors.New("invalid request sent to the board")
	case "Internet not available":
		return ErrorState, errors.New("internet not available, check your network connection")
	case "HW Error connectivity module":
		return ErrorState, errors.New("hardware error in connectivity module, check the board")
	case "HW Connectivity Module stopped":
		return ErrorState, errors.New("hardware connectivity module stopped, restart the board and check your sketch")
	case "Error initializing secure element":
		return ErrorState, errors.New("error initializing secure element, check the board and try again")
	case "Error configuring secure element":
		return ErrorState, errors.New("error configuring secure element, check the board and try again")
	case "Error locking secure element":
		return ErrorState, errors.New("error locking secure element, check the board and try again")
	case "Error generating UHWID":
		return ErrorState, errors.New("error generating UHWID, check the board and try again")
	case "Error storage begin module":
		return ErrorState, errors.New("error beginning storage module, check the board storage partitioning and try again")
	case "Fail to partition the storage":
		return ErrorState, errors.New("failed to partition the storage, check the board storage and try again")
	case "Generic error":
		return ErrorState, errors.New("generic error, check the board and try again")
	default:
		return ErrorState, errors.New("generic error, check the board and try again")
	}

}
