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
	"errors"
	"fmt"

	"github.com/arduino/arduino-cloud-cli/arduino/cli"
	"github.com/arduino/arduino-cloud-cli/internal/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	"github.com/sirupsen/logrus"
)

// CreateParams contains the parameters needed
// to find the device to be provisioned.
type CreateParams struct {
	Name string  // Device name
	Port *string // Serial port - Optional - If omitted then each serial port is analyzed
	Fqbn *string // Board FQBN - Optional - If omitted then the first device found gets selected
}

// Create command is used to provision a new arduino device
// and to add it to Arduino IoT Cloud.
func Create(params *CreateParams) (*DeviceInfo, error) {
	comm, err := cli.NewCommander()
	if err != nil {
		return nil, err
	}

	ports, err := comm.BoardList()
	if err != nil {
		return nil, err
	}
	board := boardFromPorts(ports, params)
	if board == nil {
		err = errors.New("no board found")
		return nil, err
	}

	if !board.isCrypto() {
		return nil, fmt.Errorf(
			"board with fqbn %s found at port %s is not a device with a supported crypto-chip.\n"+
				"Try the 'create-lora' command instead if it's a LoRa device"+
				" or 'create-generic' otherwise.",
			board.fqbn,
			board.port,
		)
	}

	conf, err := config.Retrieve()
	if err != nil {
		return nil, err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return nil, err
	}

	logrus.Info("Creating a new device on the cloud")
	dev, err := iotClient.DeviceCreate(board.fqbn, params.Name, board.serial, board.dType)
	if err != nil {
		return nil, err
	}

	prov := &provision{
		Commander: comm,
		Client:    iotClient,
		board:     board,
		id:        dev.Id,
	}
	if err = prov.run(); err != nil {
		if errDel := iotClient.DeviceDelete(dev.Id); errDel != nil {
			return nil, fmt.Errorf(
				"device was NOT successfully provisioned but " +
					"now we can't delete it from the cloud - please check " +
					"it on the web application.\n\nProvision error: " + err.Error() +
					"\nDeletion error: " + errDel.Error(),
			)
		}
		return nil, fmt.Errorf("cannot provision device: %w", err)
	}

	devInfo := &DeviceInfo{
		Name:   dev.Name,
		ID:     dev.Id,
		Board:  dev.Type,
		Serial: dev.Serial,
		FQBN:   dev.Fqbn,
	}
	return devInfo, nil
}
