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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/arduino/arduino-cloud-cli/arduino"
	"github.com/arduino/arduino-cloud-cli/config"
	configurationprotocol "github.com/arduino/arduino-cloud-cli/internal/board-protocols/configuration-protocol"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/configuration-protocol/cborcoders"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/transport"
	"github.com/arduino/arduino-cloud-cli/internal/boardpids"
	iotapiraw "github.com/arduino/arduino-cloud-cli/internal/iot-api-raw"
	provisioningapi "github.com/arduino/arduino-cloud-cli/internal/provisioning-api"
	"github.com/arduino/go-paths-helper"
	"github.com/sirupsen/logrus"
)

var connectionTypeIDByName = map[string]int32{
	"wifi":     1,
	"eth":      2,
	"nb":       3,
	"gsm":      4,
	"lora":     5,
	"catm1":    6,
	"cellular": 7,
}

type ConnectedBoardInfos struct {
	UHWID         string
	PublicKey     string
	Signature     string
	BLEMacAddress string
}

type ProvisioningV2BoardParams struct {
	fqbn                 string
	address              string
	protocol             string
	serial               string
	minProvSketchVersion string
	minWiFiVersion       *string
	name                 string
	connectionType       string
	netConfig            NetConfig
}

type ProvisionV2 struct {
	arduino.Commander
	iotApiClient        *iotapiraw.IoTApiRawClient
	provisioningClient  *provisioningapi.ProvisioningApiClient
	provProt            *configurationprotocol.NetworkConfigurationProtocol
	state               ConfigStatus
	configStates        *ConfigurationStates
	connectedBoardInfos ConnectedBoardInfos
	provisioningId      string
	deviceId            string
}

func NewProvisionV2(iotClient *iotapiraw.IoTApiRawClient, credentials *config.Credentials, extInterface transport.TransportInterface) *ProvisionV2 {
	provProt := configurationprotocol.NewNetworkConfigurationProtocol(&extInterface)
	return &ProvisionV2{
		iotApiClient:       iotClient,
		provisioningClient: provisioningapi.NewClient(credentials),
		provProt:           provProt,
		configStates:       NewConfigurationStates(provProt),
		deviceId:           "",
	}
}

func (p *ProvisionV2) connectToBoard(address string) error {
	err := p.provProt.Connect(address)
	return err
}

/*
 * The function return the Arduino Cloud Device ID of the new created board
 * if the process ends successfully. Otherwise, an error
 */
func (p *ProvisionV2) GetProvisioningResult() (string, error) {
	if p.deviceId == "" {
		return "", errors.New("device not provisioned")
	}
	return p.deviceId, nil
}

func (p *ProvisionV2) Run(ctx context.Context, params ProvisioningV2BoardParams) error {
	var err error
	if err = p.connectToBoard(params.address); err != nil {
		return err
	}
	p.state = WaitForConnection
	var nextState ConfigStatus

	for p.state != End {

		switch p.state {
		case WaitForConnection:
			nextState, err = p.configStates.WaitForConnection()
			if err != nil {
				nextState = End
			}
			p.state = nextState
		case WaitingForInitialStatus:
			nextState, err = p.configStates.WaitingForInitialStatus()
			if err != nil {
				nextState = End
			}
			p.state = nextState
		case WaitingForNetworkOptions:
			nextState, err = p.configStates.WaitingForNetworkOptions()
			if err != nil {
				nextState = FlashProvisioningSketch
			}
			p.state = nextState
		case BoardReady:
			p.state = GetSketchVersionRequest
		case GetSketchVersionRequest:
			err = p.getSketchVersionRequest()
			if err != nil {
				p.state = End
			}
		case WaitingSketchVersion:
			err = p.waitingSketchVersion(params.minProvSketchVersion)
			if err != nil {
				p.state = End
			}
		case FlashProvisioningSketch:
			err = p.flashProvisioningSketch(ctx, params.fqbn, params.address, params.protocol)
			if err != nil {
				p.state = End
			}
		case WiFiFWVersionRequest:
			err = p.getWiFiFWVersionRequest()
			if err != nil {
				p.state = End
			}
		case WaitingWiFiFWVersion:
			err = p.waitWiFiFWVersion(params.minWiFiVersion)
			if err != nil {
				p.state = End
			}
		case RequestBLEMAC:
			err = p.getBLEMACRequest()
			if err != nil {
				p.state = End
			}
		case WaitBLEMAC:
			err = p.waitBLEMac()
			if err != nil {
				p.state = End
			}
		case SendInitialTS:
			err = p.sendInitialTS()
			if err != nil {
				p.state = End
			}

		case IDRequest:
			err = p.getIDRequest()
			if err != nil {
				p.state = End
			}
		case WaitingPublicKey:
			err = p.waitingPublicKey()
			if err != nil {
				p.state = End
			}
		case WaitingID:
			err = p.waitingUHWID()
			if err != nil {
				p.state = End
			}
		case WaitingSignature:
			err = p.waitingSignature()
			if err != nil {
				p.state = End
			}
		case ClaimDevice:
			err = p.claimDevice(params.name, params.connectionType)
			if err != nil {
				p.state = End
			}
		case RegisterDevice:
			err = p.registerDevice(params.fqbn, params.serial)
			if err != nil {
				p.state = End
			}

		case RequestReset:
			err = p.resetBoardRequest()
			if err != nil {
				p.state = UnclaimDevice
			}
		case WaitResetResponse:
			err = p.waitingForResetResult()
			if err != nil {
				p.state = UnclaimDevice
			}
		case ConfigureNetwork:
			nextState, err = p.configStates.ConfigureNetwork(ctx, &params.netConfig)
			if err != nil {
				nextState = UnclaimDevice
			}
		case SendConnectionRequest:
			nextState, err = p.configStates.SendConnectionRequest()
			if err != nil {
				nextState = UnclaimDevice
			}
			p.state = nextState
		case WaitingForConnectionCommandResult:
			nextState, err = p.configStates.WaitingForConnectionCommandResult()
			if err != nil {
				nextState = UnclaimDevice
			}

			if nextState == MissingParameter {
				nextState = ConfigureNetwork
			}
			p.state = nextState
		case WaitingForNetworkConfigResult:
			_, err = p.configStates.WaitingForNetworkConfigResult()
			if err != nil {
				nextState = UnclaimDevice
			}
			p.state = WaitingForProvisioningResult
		case WaitingForProvisioningResult:
			err = p.waitProvisioningResult(ctx)
			if err != nil {
				p.state = UnclaimDevice
			}
		case UnclaimDevice:
			err = p.unclaimDevice()

		}

	}

	p.provProt.Close()
	return err
}

func (p *ProvisionV2) getSketchVersionRequest() error {
	logrus.Info("Provisioning V2: Requesting Sketch Version")
	getSketchVersionMessage := cborcoders.From(cborcoders.ProvisioningCommandsMessage{Command: configurationprotocol.Commands["GetSketchVersion"]})
	err := p.provProt.SendData(getSketchVersionMessage)
	if err != nil {
		return err
	}
	p.state = WaitingSketchVersion
	return nil
}

/*
 * This function returns
 * - <0 if version1  < version2
 * - =0 if version1 == version2
 * - >0 if version1  > version2
 */
func (p *ProvisionV2) compareVersions(version1, version2 string) int {
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

func (p *ProvisionV2) waitingSketchVersion(minSketchVersion string) error {
	res, err := p.provProt.ReceiveData(60)
	if err != nil {
		return err
	}

	if res == nil {
		logrus.Error("Provisioning V2: Requesting sketch Version failed, flashing...")
		p.state = FlashProvisioningSketch
		return nil
	}

	if res.Type() == cborcoders.ProvisioningSketchVersionMessageType {
		sketch_version := res.ToProvisioningSketchVersionMessage().ProvisioningSketchVersion
		logrus.Info("Provisioning V2: Received Sketch Version %s", sketch_version)

		if p.compareVersions(sketch_version, minSketchVersion) < 0 {
			logrus.Info("Provisioning V2: Sketch version %s is lower than required minimum %s. Updating...", sketch_version, minSketchVersion)
			p.state = FlashProvisioningSketch
			return nil
		}

		p.state = WiFiFWVersionRequest
	} else if res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		newState, err := p.configStates.HandleStatusMessage(status.Status)
		if err != nil {
			return err
		}
		if newState != NoneState {
			p.state = newState
			return nil
		}
	}

	return nil
}

func (p *ProvisionV2) flashProvisioningSketch(ctx context.Context, fqbn, address, protocol string) error {
	logrus.Info("Provisioning V2: Downloading provisioning sketch")
	path := paths.TempDir().Join("cloud-cli").Join("provisioning_v2_sketch")

	file, err := p.iotApiClient.DownloadProvisioningV2Sketch(fqbn, path, nil)
	if err != nil {
		logrus.Error("Provisioning V2: Downloading provisioning sketch failed")
		return err
	}

	// Try to upload the provisioning sketch
	logrus.Infof("%s\n", "Uploading provisioning sketch on the board")
	errMsg := "Provisioning V2: error while uploading the provisioning sketch"
	err = retry(ctx, 5, time.Millisecond*1000, errMsg, func() error {
		return p.UploadBin(ctx, fqbn, file, address, protocol)
	})
	if err != nil {
		return err
	}

	logrus.Info("Provisioning V2: Uploading provisioning sketch succeeded")
	sleepCtx(ctx, 3*time.Second)
	if err = p.connectToBoard(address); err != nil {
		return err
	}
	p.state = WaitForConnection
	return nil

}

func (p *ProvisionV2) getWiFiFWVersionRequest() error {
	logrus.Info("Provisioning V2: Requesting WiFi FW Version")
	getWiFiFWVersionMessage := cborcoders.From(cborcoders.ProvisioningCommandsMessage{Command: configurationprotocol.Commands["GetWiFiFWVersion"]})
	err := p.provProt.SendData(getWiFiFWVersionMessage)
	if err != nil {
		return err
	}
	p.state = WaitingWiFiFWVersion
	time.Sleep(1 * time.Second)
	return nil
}

func (p *ProvisionV2) waitWiFiFWVersion(minWiFiVersion *string) error {
	res, err := p.provProt.ReceiveData(60)
	if err != nil {
		return err
	}

	if res == nil {
		return errors.New("Provisioning V2: Requesting WiFi FW Version failed")
	}

	if res.Type() == cborcoders.ProvisioningWiFiFWVersionMessageType {
		wifi_version := res.ToProvisioningWiFiFWVersionMessage().WiFiFWVersion
		fmt.Printf("Received WiFi FW Version: %s\n", wifi_version)
		if minWiFiVersion != nil &&
			p.compareVersions(wifi_version, *minWiFiVersion) < 0 {
			return fmt.Errorf("Provisioning V2: WiFi FW version %s is lower than required minimum %s. Please update the board firmware using Arduino IDE or Arduino CLI", wifi_version, *minWiFiVersion)
		}
		p.state = RequestBLEMAC

	} else if res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		newState, err := p.configStates.HandleStatusMessage(status.Status)
		if err != nil {
			return err
		}
		if newState != NoneState {
			p.state = newState
			return nil
		}
	}
	return errors.New("Provisioning V2: WiFi FW version not received")
}

func (p *ProvisionV2) getBLEMACRequest() error {
	logrus.Info("Provisioning V2: Requesting BLE MAC")
	getblemacMessage := cborcoders.From(cborcoders.ProvisioningCommandsMessage{Command: configurationprotocol.Commands["GetBLEMac"]})
	err := p.provProt.SendData(getblemacMessage)
	if err != nil {
		return err
	}
	p.state = WaitBLEMAC
	time.Sleep(1 * time.Second)
	return nil
}

func (p *ProvisionV2) waitBLEMac() error {
	res, err := p.provProt.ReceiveData(60)
	if err != nil {
		return err
	}

	if res == nil {
		return errors.New("Provisioning V2: BLEMac was not received")
	}

	if res.Type() == cborcoders.ProvisioningBLEMacAddressMessageType {
		mac := res.ToProvisioningBLEMacAddressMessage().BLEMacAddress
		logrus.Info("Provisioning V2: Received MAC in hex: %02X\n", mac)
		macStr := fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
		p.connectedBoardInfos.BLEMacAddress = macStr
		p.state = SendInitialTS
	} else if res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		newState, err := p.configStates.HandleStatusMessage(status.Status)
		if err != nil {
			return err
		}
		if newState != NoneState {
			p.state = newState
			return nil
		}
	}
	return errors.New("Provisioning V2: BLE MAC address not received")
}

func (p *ProvisionV2) sendInitialTS() error {
	logrus.Info("Provisioning V2: Sending initial timestamp")
	ts := time.Now().Unix()
	logrus.Infof("Provisioning V2: Sending timestamp: %d\n", ts)
	tsMessage := cborcoders.From(cborcoders.ProvisioningTimestampMessage{Timestamp: uint64(ts)})
	err := p.provProt.SendData(tsMessage)
	if err != nil {
		return err
	}
	p.state = IDRequest
	time.Sleep(1 * time.Second)
	return nil
}

func (p *ProvisionV2) getIDRequest() error {
	logrus.Info("Provisioning V2: Requesting UniqueID")
	getuuidMessage := cborcoders.From(cborcoders.ProvisioningCommandsMessage{Command: configurationprotocol.Commands["GetID"]})
	err := p.provProt.SendData(getuuidMessage)
	if err != nil {
		return err
	}
	p.state = WaitingPublicKey
	return nil
}

func (p *ProvisionV2) waitingPublicKey() error {
	res, err := p.provProt.ReceiveData(60)
	if err != nil {
		return err
	}

	if res == nil {
		return errors.New("Provisioning V2: public key was not received")
	}

	if res.Type() == cborcoders.ProvisioningPublicKeyMessageType {
		pubKey := res.ToProvisioningPublicKeyMessage().ProvisioningPublicKey
		logrus.Infof("Provisioning V2: Received Public Key\n")
		p.connectedBoardInfos.PublicKey = pubKey
		p.state = WaitingID
	} else if res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		newState, err := p.configStates.HandleStatusMessage(status.Status)
		if err != nil {
			return err
		}
		if newState == MissingParameter {
			p.state = SendInitialTS
			return nil
		}

		if newState != NoneState {
			p.state = newState
			return nil
		}
	}
	return errors.New("Provisioning V2: Public Key not received")
}

func (p *ProvisionV2) waitingUHWID() error {
	res, err := p.provProt.ReceiveData(60)
	if err != nil {
		return err
	}

	if res == nil {
		return errors.New("Provisioning V2: UniqueID was not received")
	}

	if res.Type() == cborcoders.ProvisioningUniqueIdMessageType {
		uhwid := res.ToProvisioningUniqueIdMessage().UniqueId
		logrus.Infof("Provisioning V2: Received UniqueID\n")
		uhwidString := string(uhwid[:])
		p.connectedBoardInfos.UHWID = uhwidString
		p.state = WaitingSignature
	} else if res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		newState, err := p.configStates.HandleStatusMessage(status.Status)
		if err != nil {
			return err
		}

		if newState != NoneState {
			p.state = newState
			return nil
		}
	}
	return errors.New("Provisioning V2: UniqueID was not received")
}

func (p *ProvisionV2) waitingSignature() error {
	res, err := p.provProt.ReceiveData(60)
	if err != nil {
		return err
	}

	if res == nil {
		return errors.New("Provisioning V2: Signature was not received")
	}

	if res.Type() == cborcoders.ProvisioningSignatureMessageType {
		signature := res.ToProvisioningSignatureMessage().Signature
		logrus.Infof("Provisioning V2: Received Signature\n")
		p.connectedBoardInfos.Signature = string(signature[:])
		p.state = ClaimDevice
	} else if res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		newState, err := p.configStates.HandleStatusMessage(status.Status)
		if err != nil {
			return err
		}

		if newState != NoneState {
			p.state = newState
			return nil
		}
	}
	return errors.New("Provisioning V2: Signature was not received")
}

func (p *ProvisionV2) claimDevice(name, connectionType string) error {
	logrus.Info("Provisioning V2: Claiming device...")

	claimData := provisioningapi.ClaimData{
		BLEMac:         p.connectedBoardInfos.BLEMacAddress,
		BoardToken:     p.connectedBoardInfos.Signature,
		ConnectionType: connectionType,
		DeviceName:     name,
	}

	provResp, provErr, err := p.provisioningClient.ClaimDevice(claimData)
	if err != nil {
		return err
	}

	if provErr != nil {
		if provErr.ErrCode == 1 || provErr.ErrCode == 2 {
			logrus.Warn("Provisioning V2: Device claim failed. The board has to migrate")
			p.state = RegisterDevice
		} else if provErr.ErrCode == 3 {
			// If the device key and the DB key are different
			return fmt.Errorf("Provisioning V2: Device claim failed. Keys do not match. Please contact the Arduino Support with this hardware id: %s", p.connectedBoardInfos.UHWID)
		} else {
			return fmt.Errorf("Provisioning V2: Device claim failed with error: %s", provErr.Err)
		}
	}

	if provResp != nil {
		p.provisioningId = provResp.OnboardId
		p.state = RequestReset
		return nil
	}

	return errors.New("Provisioning V2: Device ID not received")
}

func (p *ProvisionV2) registerDevice(fqbn, serial string) error {
	logrus.Info("Provisioning V2: Registering device...")

	registerData := provisioningapi.RegisterBoardData{
		PID:              boardpids.ArduinoFqbnToPID[fqbn],
		PublicKey:        p.connectedBoardInfos.PublicKey,
		Serial:           &serial,
		UniqueHardwareID: p.connectedBoardInfos.UHWID,
		VID:              boardpids.ArduinoVendorID, //Only Arduino boards can support Provisioning 2.0
	}

	provErr, err := p.provisioningClient.RegisterDevice(registerData)
	if err != nil {
		return err
	}

	if provErr != nil {
		return fmt.Errorf("Provisioning V2: Device registration failed with error: %s", provErr.Err)
	}

	logrus.Info("Provisioning V2: Device registered successfully, claiming...")
	p.state = ClaimDevice
	return nil
}

func (p *ProvisionV2) resetBoardRequest() error {
	logrus.Info("Provisioning V2: Requesting Reset Stored Credentials")
	resetMessage := cborcoders.From(cborcoders.ProvisioningCommandsMessage{Command: configurationprotocol.Commands["Reset"]})
	err := p.provProt.SendData(resetMessage)
	if err != nil {
		return err
	}
	p.state = WaitResetResponse
	return nil
}

func (p *ProvisionV2) waitingForResetResult() error {
	res, err := p.provProt.ReceiveData(60)
	if err != nil {
		return err
	}

	if res != nil && res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		if status.Status == 4 {
			logrus.Info("Provisioning V2: Reset Stored Credentials successful")
			p.state = ConfigureNetwork
		} else {
			newState, err := p.configStates.HandleStatusMessage(status.Status)
			if err != nil {
				return err
			}

			if newState != NoneState {
				p.state = newState
				return nil
			}
		}
	}

	return errors.New("Provisioning V2: Reset Stored Credentials failed")
}

func (p *ProvisionV2) waitProvisioningResult(ctx context.Context) error {
	logrus.Info("Provisioning V2: Waiting for provisioning result...")

	for n := 0; n < 20; n++ {
		res, err := p.provisioningClient.GetProvisioningDetail(p.provisioningId)
		if err != nil {
			return err
		}
		if res.DeviceID != nil {
			p.deviceId = *res.DeviceID
			p.state = End
		}
		sleepCtx(ctx, 10*time.Second)
	}
	return errors.New("Provisioning V2: Timeout expires for board provisioning. The board was not able to reach the Arduino IoT Cloud for completing the provisioning.")
}

func (p *ProvisionV2) unclaimDevice() error {
	logrus.Warnf("Provisioning V2: Something went wrong, unclaiming device...")
	_, err := p.provisioningClient.UnclaimDevice(p.provisioningId)
	p.state = End
	return err
}
