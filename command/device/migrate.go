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
	"errors"
	"fmt"
	"time"

	"github.com/arduino/arduino-cloud-cli/arduino"
	"github.com/arduino/arduino-cloud-cli/arduino/cli"
	"github.com/arduino/arduino-cloud-cli/config"
	configurationprotocol "github.com/arduino/arduino-cloud-cli/internal/board-protocols/configuration-protocol"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/transport"
	iotapiraw "github.com/arduino/arduino-cloud-cli/internal/iot-api-raw"
	"github.com/arduino/arduino-cloud-cli/internal/serial"
	"github.com/sirupsen/logrus"
)

// CreateParams contains the parameters needed
// to find the device to be provisioned.
type MigrateParams struct {
	Port *string // Serial port - Optional - If omitted then each serial port is analyzed
	FQBN *string // Board FQBN - Optional - If omitted then the first device found gets selected
}

// Create command is used to provision a new arduino device
// and to add it to Arduino IoT Cloud.
func Migrate(ctx context.Context, params *MigrateParams, cred *config.Credentials) error {
	comm, err := cli.NewCommander()
	if err != nil {
		return err
	}

	ports, err := comm.BoardList(ctx)
	if err != nil {
		return err
	}
	board := boardFromPorts(ports, &CreateParams{
		Port: params.Port,
		FQBN: params.FQBN,
	})
	if board == nil {
		err = errors.New("no board found")
		return err
	}

	iotApiRawClient := iotapiraw.NewClient(cred)

	boardProvisioningDetails, err := iotApiRawClient.GetBoardDetailByFQBN(board.fqbn)
	if err != nil {
		return err
	}

	if boardProvisioningDetails.Provisioning != nil && *boardProvisioningDetails.Provisioning != "v2" {
		logrus.Info("The detected board is not compatible with Bluetooth provisioning")
		return fmt.Errorf("the board with fqbn %s found at port %s is not compatible with Bluetooth provisioning", board.fqbn, board.address)

	}
	logrus.Info("Board compatible with Bluetooth provisioning detected")
	return runMigrateCommand(ctx, &comm, iotApiRawClient, board, boardProvisioningDetails)
}

func runMigrateCommand(ctx context.Context, comm *arduino.Commander, iotApiRawClient *iotapiraw.IoTApiRawClient, board *board, boardProvisioningDetails *iotapiraw.BoardType) error {
	fwFlasher := NewProvisioningV2SketchFlasher(comm, iotApiRawClient)

	var extInterface transport.TransportInterface
	extInterface = &serial.Serial{}
	provProt := configurationprotocol.NewNetworkConfigurationProtocol(&extInterface)
	confStates := NewConfigurationStates(provProt)

	logrus.Infof("Flashing provisioning sketch to enable Bluetooth provisioning")

	err := fwFlasher.FlashProvisioningV2Sketch(ctx, board.fqbn, board.address, board.protocol)
	if err != nil {
		return err
	}

	logrus.Info("Uploading provisioning sketch succeeded, waiting for board to be ready")
	sleepCtx(ctx, 10*time.Second)

	err = provProt.Connect(board.address)
	if err != nil {
		return err
	}
	defer provProt.Close()

	state := WaitForConnection
	nextState := NoneState
	for state != End && state != ErrorState {
		switch state {
		case WaitForConnection:
			nextState, err = confStates.WaitForConnection()
		case WaitingForInitialStatus:
			nextState, err = confStates.WaitingForInitialStatus()
		case WaitingForNetworkOptions:
			nextState, err = confStates.WaitingForNetworkOptions()
		case BoardReady:
			nextState = WiFiFWVersionRequest
		case WiFiFWVersionRequest:
			nextState, err = confStates.GetWiFiFWVersionRequest(ctx)
		case WaitingWiFiFWVersion:
			_, err = confStates.WaitWiFiFWVersion(boardProvisioningDetails.MinWiFiVersion)
			if err == nil {
				nextState = End
			}
		}

		if nextState != NoneState {
			state = nextState
		}
	}

	return err
}
