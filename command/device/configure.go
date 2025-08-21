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

type NetworkConfigure struct {
	configStates   *ConfigurationStates
	state          ConfigStatus
	configProtocol *configurationprotocol.NetworkConfigurationProtocol
}

func NewNetworkConfigure(extInterface transport.TransportInterface, configProtocol *configurationprotocol.NetworkConfigurationProtocol) *NetworkConfigure {
	return &NetworkConfigure{
		configStates:   NewConfigurationStates(extInterface, configProtocol),
		configProtocol: configProtocol,
	}
}

func (nc *NetworkConfigure) Run(ctx context.Context, netConfig *NetConfig) error {
	nc.state = WaitForConnection
	var err error
	var nextState ConfigStatus
	for nc.state != End {

		switch nc.state {
		case WaitForConnection:
			nextState, err = nc.configStates.WaitForConnection()
			if err != nil {
				nextState = End
			}
			nc.state = nextState
		case WaitingForInitialStatus:
			nextState, err = nc.configStates.WaitingForInitialStatus()
			if err != nil {
				nextState = End
			}
			nc.state = nextState
		case WaitingForNetworkOptions:
			nextState, err = nc.configStates.WaitingForNetworkOptions()
			if err != nil {
				nextState = End
			}
			nc.state = nextState
		case BoardReady:
			nc.state = ConfigureNetwork
		case ConfigureNetwork:
			nextState, err = nc.configStates.ConfigureNetwork(ctx, netConfig)
			if err != nil {
				nextState = End
			}
			nc.state = nextState
		case SendConnectionRequest:
			nextState, err = nc.configStates.SendConnectionRequest()
			if err != nil {
				nextState = End
			}
			nc.state = nextState
		case WaitingForConnectionCommandResult:
			nextState, err = nc.configStates.WaitingForConnectionCommandResult()
			if err != nil {
				nextState = End
			}
			nc.state = nextState
		case MissingParameter:
			nc.state = ConfigureNetwork
		case WaitingForNetworkConfigResult:
			nextState, err = nc.configStates.WaitingForNetworkConfigResult()
			if err != nil {
				nextState = End
			}
			nc.state = nextState
		}

	}

	nc.configProtocol.Close()
	return err
}
