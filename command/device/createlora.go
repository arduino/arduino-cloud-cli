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
	"time"

	"github.com/arduino/arduino-cloud-cli/arduino/cli"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	iotclient "github.com/arduino/iot-client-go"
	"github.com/sirupsen/logrus"
	"go.bug.st/serial"
)

const (
	deveuiUploadAttempts = 3
	deveuiUploadWait     = 1000

	serialEUIAttempts = 4
	serialEUIWait     = 2000
	serialEUITimeout  = 3500
	serialEUIBaudrate = 9600

	// dev-eui is an IEEE EUI64 address, so it must have length of 8 bytes.
	// It's retrieved as hexadecimal string, thus 16 chars are expected.
	deveuiLength = 16
)

// DeviceLoraInfo contains the most interesting
// parameters of an Arduino IoT Cloud LoRa device.
type DeviceLoraInfo struct {
	DeviceInfo
	AppEUI string `json:"app_eui"`
	AppKey string `json:"app_key"`
	EUI    string `json:"eui"`
}

// CreateLoRaParams contains the parameters needed
// to provision a LoRa device.
type CreateLoraParams struct {
	CreateParams
	FrequencyPlan string
}

// CreateLora command is used to provision a new LoRa arduino device
// and to add it to Arduino IoT Cloud.
func CreateLora(ctx context.Context, params *CreateLoraParams, cred *config.Credentials) (*DeviceLoraInfo, error) {
	comm, err := cli.NewCommander()
	if err != nil {
		return nil, err
	}

	ports, err := comm.BoardList(ctx)
	if err != nil {
		return nil, err
	}
	board := boardFromPorts(ports, &params.CreateParams)
	if board == nil {
		err = errors.New("no board found")
		return nil, err
	}

	if !board.isLora() {
		return nil, fmt.Errorf(
			"board with fqbn %s found at port %s is not a LoRa device."+
				" Try the 'create' command instead if it's a device with a supported crypto-chip"+
				" or 'create-generic' otherwise",
			board.fqbn,
			board.address,
		)
	}

	bin, err := downloadProvisioningFile(ctx, board.fqbn)
	if err != nil {
		return nil, err
	}

	logrus.Infof("%s", "Uploading deveui sketch on the LoRa board")
	errMsg := "Error while uploading the LoRa provisioning binary"
	err = retry(ctx, deveuiUploadAttempts, deveuiUploadWait*time.Millisecond, errMsg, func() error {
		return comm.UploadBin(ctx, board.fqbn, bin, board.address, board.protocol)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload LoRa provisioning binary: %w", err)
	}

	eui, err := extractEUI(ctx, board.address)
	if err != nil {
		return nil, err
	}

	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return nil, err
	}

	logrus.Info("Creating a new device on the cloud")
	dev, err := iotClient.DeviceLoraCreate(ctx, params.Name, board.serial, board.dType, eui, params.FrequencyPlan)
	if err != nil {
		return nil, err
	}

	devInfo, err := getDeviceLoraInfo(ctx, iotClient, dev)
	if err != nil {
		// Don't use the passed context for the cleanup because it could be cancelled.
		errDel := iotClient.DeviceDelete(context.Background(), dev.DeviceId)
		if errDel != nil { // Oh no
			return nil, fmt.Errorf(
				"device was successfully provisioned and configured on IoT-API but " +
					"now we can't fetch its information nor delete it - please check " +
					"it on the web application.\n\nFetch error: " + err.Error() +
					"\nDeletion error: " + errDel.Error(),
			)
		}
		return nil, fmt.Errorf("%s: %w", "cannot provision LoRa device", err)
	}
	return devInfo, nil
}

// extractEUI extracts the EUI from the provisioned lora board.
func extractEUI(ctx context.Context, port string) (string, error) {
	var ser serial.Port

	logrus.Infof("%s\n", "Connecting to the board through serial port")
	errMsg := "Error while connecting to the board"
	err := retry(ctx, serialEUIAttempts, serialEUIWait*time.Millisecond, errMsg, func() error {
		var err error
		ser, err = serial.Open(port, &serial.Mode{BaudRate: serialEUIBaudrate})
		return err
	})
	if err != nil {
		return "", fmt.Errorf("failed to extract deveui from the board: %w", err)
	}

	err = ser.SetReadTimeout(serialEUITimeout * time.Millisecond)
	if err != nil {
		return "", fmt.Errorf("setting serial read timeout: %w", err)
	}

	buff := make([]byte, deveuiLength)
	n, err := ser.Read(buff)
	if err != nil {
		return "", fmt.Errorf("reading from serial: %w", err)
	}

	if n < deveuiLength {
		return "", errors.New("cannot read eui from the device")
	}
	eui := string(buff)
	return eui, nil
}

func getDeviceLoraInfo(ctx context.Context, iotClient *iot.Client, loraDev *iotclient.ArduinoLoradevicev1) (*DeviceLoraInfo, error) {
	dev, err := iotClient.DeviceShow(ctx, loraDev.DeviceId)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve device from the cloud: %w", err)
	}

	devInfo := &DeviceLoraInfo{
		DeviceInfo: DeviceInfo{
			Name:   dev.Name,
			ID:     dev.Id,
			Board:  dev.Type,
			Serial: dev.Serial,
			FQBN:   dev.Fqbn,
		},
		AppEUI: loraDev.AppEui,
		AppKey: loraDev.AppKey,
		EUI:    loraDev.Eui,
	}
	return devInfo, nil
}
