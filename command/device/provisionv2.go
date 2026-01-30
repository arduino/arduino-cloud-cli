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
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/arduino/arduino-cloud-cli/arduino"
	"github.com/arduino/arduino-cloud-cli/config"
	configurationprotocol "github.com/arduino/arduino-cloud-cli/internal/board-protocols/configuration-protocol"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/configuration-protocol/cborcoders"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/transport"
	"github.com/arduino/arduino-cloud-cli/internal/boardpids"
	iotapiraw "github.com/arduino/arduino-cloud-cli/internal/iot-api-raw"
	provisioningapi "github.com/arduino/arduino-cloud-cli/internal/provisioning-api"
	"github.com/arduino/go-paths-helper"
	"github.com/beevik/ntp"
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

const (
	MaxRetriesFlashProvSketch    = 5
	MaxRetriesProvisioningResult = 20
)

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
	FWFlasher           *ProvisioningV2SketchFlasher
	iotApiClient        *iotapiraw.IoTApiRawClient
	provisioningClient  *provisioningapi.ProvisioningApiClient
	provProt            *configurationprotocol.NetworkConfigurationProtocol
	configStates        *ConfigurationStates
	connectedBoardInfos ConnectedBoardInfos
	provisioningId      string
	deviceId            string
}

func NewProvisionV2(comm *arduino.Commander, iotClient *iotapiraw.IoTApiRawClient, credentials *config.Credentials, extInterface transport.TransportInterface) *ProvisionV2 {
	provProt := configurationprotocol.NewNetworkConfigurationProtocol(&extInterface)
	return &ProvisionV2{
		FWFlasher:          NewProvisioningV2SketchFlasher(comm, iotClient),
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
	state := WaitForConnection
	nextState := NoneState

	// FSM for Provisioning 2.0
	for state != End && state != ErrorState {

		switch state {
		case WaitForConnection:
			nextState, err = p.configStates.WaitForConnection()
		case WaitingForInitialStatus:
			nextState, err = p.configStates.WaitingForInitialStatus()
			if err != nil {
				nextState = FlashProvisioningSketch
			}
		case WaitingForNetworkOptions:
			nextState, err = p.configStates.WaitingForNetworkOptions()
			if err != nil {
				nextState = FlashProvisioningSketch
			}
		case BoardReady:
			nextState = GetSketchVersionRequest
		case GetSketchVersionRequest:
			nextState, err = p.getSketchVersionRequest()
		case WaitingSketchVersion:
			nextState, err = p.waitingSketchVersion(params.minProvSketchVersion)
		case FlashProvisioningSketch:
			nextState, err = p.flashProvisioningSketch(ctx, params.fqbn, params.address, params.protocol)
		case WiFiFWVersionRequest:
			nextState, err = p.getWiFiFWVersionRequest(ctx)
		case WaitingWiFiFWVersion:
			nextState, err = p.waitWiFiFWVersion(params.minWiFiVersion)
		case RequestBLEMAC:
			nextState, err = p.getBLEMACRequest(ctx)
		case WaitBLEMAC:
			nextState, err = p.waitBLEMac()
		case SendInitialTS:
			nextState, err = p.sendInitialTS(ctx)
		case IDRequest:
			nextState, err = p.getIDRequest()
		case WaitingPublicKey:
			nextState, err = p.waitingPublicKey()
		case WaitingID:
			nextState, err = p.waitingUHWID()
		case WaitingSignature:
			nextState, err = p.waitingSignature()
		case ClaimDevice:
			nextState, err = p.claimDevice(params.name, params.connectionType)
		case RegisterDevice:
			nextState, err = p.registerDevice(params.fqbn, params.serial)
		case RequestReset:
			nextState, err = p.resetBoardRequest()
			if err != nil {
				nextState = UnclaimDevice
			}
		case WaitResetResponse:
			nextState, err = p.waitingForResetResult()
			if err != nil {
				nextState = UnclaimDevice
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
		case WaitingForConnectionCommandResult:
			nextState, err = p.configStates.WaitingForConnectionCommandResult()
			if err != nil {
				nextState = UnclaimDevice
			}

			if nextState == MissingParameter {
				nextState = ConfigureNetwork
			}
		case WaitingForNetworkConfigResult:
			_, err = p.configStates.WaitingForNetworkConfigResult()
			if err != nil {
				nextState = UnclaimDevice
			} else {
				nextState = WaitingForProvisioningResult
			}
		case WaitingForProvisioningResult:
			nextState, err = p.waitProvisioningResult(ctx)
			if err != nil {
				nextState = UnclaimDevice
			}
		case UnclaimDevice:
			nextState, _ = p.unclaimDevice()
		}

		if nextState != NoneState {
			state = nextState
		}

	}

	p.provProt.Close()
	return err
}

func (p *ProvisionV2) getSketchVersionRequest() (ConfigStatus, error) {
	logrus.Info("Provisioning V2: Requesting Sketch Version")
	getSketchVersionMessage := cborcoders.From(cborcoders.ProvisioningCommandsMessage{Command: configurationprotocol.Commands["GetSketchVersion"]})
	err := p.provProt.SendData(getSketchVersionMessage)
	if err != nil {
		return ErrorState, err
	}

	return WaitingSketchVersion, nil
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

func (p *ProvisionV2) waitingSketchVersion(minSketchVersion string) (ConfigStatus, error) {
	res, err := p.provProt.ReceiveData(CommandResponseTimeoutLong_s)
	if err != nil {
		return ErrorState, err
	}

	if res == nil {
		logrus.Error("Provisioning V2: Requesting sketch Version failed, flashing...")
		return FlashProvisioningSketch, nil
	}

	if res.Type() == cborcoders.ProvisioningSketchVersionMessageType {
		sketch_version := res.ToProvisioningSketchVersionMessage().ProvisioningSketchVersion
		logrus.Infof("Provisioning V2: Received Sketch Version %s", sketch_version)

		if p.compareVersions(sketch_version, minSketchVersion) < 0 {
			logrus.Infof("Provisioning V2: Sketch version %s is lower than required minimum %s. Updating...", sketch_version, minSketchVersion)
			return FlashProvisioningSketch, nil
		}

		return WiFiFWVersionRequest, nil
	}

	if res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		if status.Status == -7 {
			return FlashProvisioningSketch, nil
		}
		return p.configStates.HandleStatusMessage(status.Status)
	}

	return NoneState, nil
}

func (p *ProvisionV2) flashProvisioningSketch(ctx context.Context, fqbn, address, protocol string) (ConfigStatus, error) {
	p.provProt.Close()
	err := p.FWFlasher.FlashProvisioningV2Sketch(ctx, fqbn, address, protocol)
	if err != nil {
		return ErrorState, err
	}

	logrus.Info("Provisioning V2: Uploading provisioning sketch succeeded")
	sleepCtx(ctx, 3*time.Second)
	if err = p.connectToBoard(address); err != nil {
		return ErrorState, err
	}

	return WaitForConnection, nil
}

func (p *ProvisionV2) getWiFiFWVersionRequest(ctx context.Context) (ConfigStatus, error) {
	logrus.Info("Provisioning V2: Requesting WiFi FW Version")
	getWiFiFWVersionMessage := cborcoders.From(cborcoders.ProvisioningCommandsMessage{Command: configurationprotocol.Commands["GetWiFiFWVersion"]})
	err := p.provProt.SendData(getWiFiFWVersionMessage)
	if err != nil {
		return ErrorState, err
	}
	sleepCtx(ctx, 1*time.Second)
	return WaitingWiFiFWVersion, nil
}

func (p *ProvisionV2) waitWiFiFWVersion(minWiFiVersion *string) (ConfigStatus, error) {
	res, err := p.provProt.ReceiveData(CommandResponseTimeoutLong_s)
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
			p.compareVersions(wifi_version, *minWiFiVersion) < 0 {
			return ErrorState, fmt.Errorf("provisioning V2: WiFi FW version %s is lower than required minimum %s. Please update the board firmware using Arduino IDE or Arduino CLI", wifi_version, *minWiFiVersion)
		}

		return RequestBLEMAC, nil
	}

	if res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		return p.configStates.HandleStatusMessage(status.Status)
	}

	return ErrorState, errors.New("provisioning V2: WiFi FW version not received")
}

func (p *ProvisionV2) getBLEMACRequest(ctx context.Context) (ConfigStatus, error) {
	logrus.Info("Provisioning V2: Requesting BLE MAC")
	getblemacMessage := cborcoders.From(cborcoders.ProvisioningCommandsMessage{Command: configurationprotocol.Commands["GetBLEMac"]})
	err := p.provProt.SendData(getblemacMessage)
	if err != nil {
		return ErrorState, err
	}
	sleepCtx(ctx, 1*time.Second)
	return WaitBLEMAC, nil
}

func (p *ProvisionV2) waitBLEMac() (ConfigStatus, error) {
	res, err := p.provProt.ReceiveData(CommandResponseTimeoutLong_s)
	if err != nil {
		return ErrorState, err
	}

	if res == nil {
		return ErrorState, errors.New("provisioning V2: BLEMac was not received")
	}

	if res.Type() == cborcoders.ProvisioningBLEMacAddressMessageType {
		mac := res.ToProvisioningBLEMacAddressMessage().BLEMacAddress
		logrus.Infof("Provisioning V2: Received MAC in hex: %02X", mac)
		macStr := fmt.Sprintf("%02X:%02X:%02X:%02X:%02X:%02X", mac[0], mac[1], mac[2], mac[3], mac[4], mac[5])
		p.connectedBoardInfos.BLEMacAddress = macStr
		return SendInitialTS, nil
	}

	if res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		return p.configStates.HandleStatusMessage(status.Status)
	}

	return ErrorState, errors.New("provisioning V2: BLE MAC address not received")
}

func (p *ProvisionV2) sendInitialTS(ctx context.Context) (ConfigStatus, error) {
	logrus.Info("Provisioning V2: Sending initial timestamp")
	var ts int64
	t, err := ntp.Time("time.arduino.cc")
	if err == nil {
		ts = t.Unix()
	} else {
		logrus.Warnf("Provisioning V2: Cannot get time from NTP server, using local time: %v", err)
		ts = time.Now().Unix()
	}

	logrus.Infof("Provisioning V2: Sending timestamp: %d", ts)
	tsMessage := cborcoders.From(cborcoders.ProvisioningTimestampMessage{Timestamp: uint64(ts)})
	err = p.provProt.SendData(tsMessage)
	if err != nil {
		return ErrorState, err
	}
	sleepCtx(ctx, 1*time.Second)
	return IDRequest, nil
}

func (p *ProvisionV2) getIDRequest() (ConfigStatus, error) {
	logrus.Info("Provisioning V2: Requesting UniqueID")
	getuuidMessage := cborcoders.From(cborcoders.ProvisioningCommandsMessage{Command: configurationprotocol.Commands["GetID"]})
	err := p.provProt.SendData(getuuidMessage)
	if err != nil {
		return ErrorState, err
	}

	return WaitingPublicKey, nil
}

func (p *ProvisionV2) waitingPublicKey() (ConfigStatus, error) {
	res, err := p.provProt.ReceiveData(CommandResponseTimeoutLong_s)
	if err != nil {
		return ErrorState, err
	}

	if res == nil {
		return ErrorState, errors.New("provisioning V2: public key was not received")
	}

	if res.Type() == cborcoders.ProvisioningPublicKeyMessageType {
		pubKey := res.ToProvisioningPublicKeyMessage().ProvisioningPublicKey
		logrus.Info("Provisioning V2: Received Public Key")
		p.connectedBoardInfos.PublicKey = pubKey
		return WaitingID, nil
	}

	if res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		newState, err := p.configStates.HandleStatusMessage(status.Status)
		if newState == MissingParameter {
			return SendInitialTS, nil
		}
		return newState, err
	}
	return ErrorState, errors.New("provisioning V2: Public Key not received")
}

func (p *ProvisionV2) waitingUHWID() (ConfigStatus, error) {
	res, err := p.provProt.ReceiveData(CommandResponseTimeoutLong_s)
	if err != nil {
		return ErrorState, err
	}

	if res == nil {
		return ErrorState, errors.New("provisioning V2: UniqueID was not received")
	}

	if res.Type() == cborcoders.ProvisioningUniqueIdMessageType {
		uhwid := res.ToProvisioningUniqueIdMessage().UniqueId
		logrus.Infof("Provisioning V2: Received UniqueID")
		uhwidString := fmt.Sprintf("%02x", uhwid)
		p.connectedBoardInfos.UHWID = uhwidString
		return WaitingSignature, nil
	}

	if res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		return p.configStates.HandleStatusMessage(status.Status)
	}

	return ErrorState, errors.New("provisioning V2: UniqueID was not received")
}

func (p *ProvisionV2) waitingSignature() (ConfigStatus, error) {
	res, err := p.provProt.ReceiveData(CommandResponseTimeoutLong_s)
	if err != nil {
		return ErrorState, err
	}

	if res == nil {
		return ErrorState, errors.New("provisioning V2: Signature was not received")
	}

	if res.Type() == cborcoders.ProvisioningSignatureMessageType {
		signature := res.ToProvisioningSignatureMessage().Signature
		logrus.Infof("Provisioning V2: Received Signature")

		signatureString := strings.TrimRightFunc(fmt.Sprintf("%s", signature), func(r rune) bool {
			return unicode.IsLetter(r) == false && unicode.IsNumber(r) == false
		})
		p.connectedBoardInfos.Signature = signatureString
		return ClaimDevice, nil
	}

	if res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		return p.configStates.HandleStatusMessage(status.Status)
	}

	return ErrorState, errors.New("provisioning V2: Signature was not received")
}

func (p *ProvisionV2) claimDevice(name, connectionType string) (ConfigStatus, error) {
	logrus.Info("Provisioning V2: Claiming device...")

	claimData := provisioningapi.ClaimData{
		BLEMac:         p.connectedBoardInfos.BLEMacAddress,
		BoardToken:     p.connectedBoardInfos.Signature,
		ConnectionType: connectionType,
		DeviceName:     name,
	}

	provResp, provErr, err := p.provisioningClient.ClaimDevice(claimData)
	if err != nil {
		return ErrorState, fmt.Errorf("provisioning V2: failed to claim device: %w", err)
	}

	if provErr != nil {
		if provErr.ErrCode == 1 || provErr.ErrCode == 2 {
			logrus.Warn("Provisioning V2: Device claim failed. The board has to migrate")
			return RegisterDevice, nil
		}

		if provErr.ErrCode == 3 {
			// If the device key and the DB key are different
			return ErrorState, fmt.Errorf("provisioning V2: Device claim failed. Keys do not match. Please contact the Arduino Support with this hardware id: %s", p.connectedBoardInfos.UHWID)
		}

		return ErrorState, fmt.Errorf("provisioning V2: Device claim failed with error: %s", provErr.Err)
	}

	if provResp != nil {
		p.provisioningId = provResp.OnboardId
		return RequestReset, nil
	}

	return ErrorState, errors.New("provisioning V2: Device ID not received")
}

func (p *ProvisionV2) registerDevice(fqbn, serial string) (ConfigStatus, error) {
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
		return ErrorState, fmt.Errorf("provisioning V2: failed to register device: %w", err)
	}

	if provErr != nil {
		return ErrorState, fmt.Errorf("provisioning V2: Device registration failed with error: %s", provErr.Err)
	}

	logrus.Info("Provisioning V2: Device registered successfully, claiming...")
	return ClaimDevice, nil
}

func (p *ProvisionV2) resetBoardRequest() (ConfigStatus, error) {
	logrus.Info("Provisioning V2: Requesting Reset Stored Credentials")
	resetMessage := cborcoders.From(cborcoders.ProvisioningCommandsMessage{Command: configurationprotocol.Commands["Reset"]})
	err := p.provProt.SendData(resetMessage)
	if err != nil {
		return ErrorState, err
	}

	return WaitResetResponse, nil
}

func (p *ProvisionV2) waitingForResetResult() (ConfigStatus, error) {
	res, err := p.provProt.ReceiveData(CommandResponseTimeoutLong_s)
	if err != nil {
		return ErrorState, err
	}

	if res != nil && res.Type() == cborcoders.ProvisioningStatusMessageType {
		status := res.ToProvisioningStatusMessage()
		if status.Status == 4 {
			logrus.Info("Provisioning V2: Reset Stored Credentials successful")
			return ConfigureNetwork, nil
		}
		return p.configStates.HandleStatusMessage(status.Status)
	}

	return ErrorState, errors.New("provisioning V2: Reset Stored Credentials failed")
}

func (p *ProvisionV2) waitProvisioningResult(ctx context.Context) (ConfigStatus, error) {
	logrus.Info("Provisioning V2: Waiting for provisioning result...")

	for n := 0; n < MaxRetriesProvisioningResult; n++ {
		res, err := p.provisioningClient.GetProvisioningDetail(p.provisioningId)
		if err != nil {
			return ErrorState, err
		}
		if res.DeviceID != nil {
			p.deviceId = *res.DeviceID
			return End, nil
		}
		sleepCtx(ctx, 10*time.Second)
	}
	return ErrorState, errors.New("provisioning V2: Timeout expires for board provisioning. The board was not able to reach the Arduino IoT Cloud for completing the provisioning")
}

func (p *ProvisionV2) unclaimDevice() (ConfigStatus, error) {
	logrus.Warnf("Provisioning V2: Something went wrong, unclaiming device...")
	_, err := p.provisioningClient.UnclaimDevice(p.provisioningId)
	return End, err
}

type ProvisioningV2SketchFlasher struct {
	arduino.Commander
	iotApiClient *iotapiraw.IoTApiRawClient
}

func NewProvisioningV2SketchFlasher(comm *arduino.Commander, iotClient *iotapiraw.IoTApiRawClient) *ProvisioningV2SketchFlasher {
	return &ProvisioningV2SketchFlasher{
		Commander:    *comm,
		iotApiClient: iotClient,
	}
}

func (sf *ProvisioningV2SketchFlasher) FlashProvisioningV2Sketch(ctx context.Context, fqbn, address, protocol string) error {
	logrus.Info("Provisioning V2: Downloading provisioning sketch")
	path := paths.TempDir().Join("cloud-cli").Join("provisioning_v2_sketch")

	file, err := sf.iotApiClient.DownloadProvisioningV2Sketch(fqbn, path, nil)
	if err != nil {
		logrus.Error("Provisioning V2: Downloading provisioning sketch failed")
		return err
	}

	// Try to upload the provisioning sketch
	logrus.Info("Provisioning V2: Uploading provisioning sketch on the board")
	errMsg := "Provisioning V2: error while uploading the provisioning sketch"
	err = retry(ctx, MaxRetriesFlashProvSketch, time.Millisecond*1000, errMsg, func() error {
		return sf.UploadBin(ctx, fqbn, file, address, protocol)
	})
	if err != nil {
		return err
	}

	err = os.Remove(file)
	if err != nil {
		logrus.Error("Provisioning V2: Removing temporary file failed")
		return err
	}
	return nil
}
