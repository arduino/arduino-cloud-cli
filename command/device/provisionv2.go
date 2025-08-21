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
	"github.com/arduino/arduino-cloud-cli/arduino"
	configurationprotocol "github.com/arduino/arduino-cloud-cli/internal/board-protocols/configuration-protocol"
	"github.com/arduino/arduino-cloud-cli/internal/board-protocols/transport"
	provisioningapi "github.com/arduino/arduino-cloud-cli/internal/provisioning-api"
)

type ProvisionV2 struct {
	arduino.Commander
	provisioningClient *provisioningapi.ProvisioningApiClient
	serial             transport.TransportInterface
	provProt           *configurationprotocol.NetworkConfigurationProtocol
	board              *board
	id                 string
}

func NewProvisionV2(provisioningClient *provisioningapi.ProvisioningApiClient, serial transport.TransportInterface, provProt *configurationprotocol.NetworkConfigurationProtocol, board *board, id string) *ProvisionV2 {
	return &ProvisionV2{
		provisioningClient: provisioningClient,
		serial:             serial,
		provProt:           provProt,
		board:              board,
		id:                 id,
	}
}

func (p *ProvisionV2) run() error {

}
