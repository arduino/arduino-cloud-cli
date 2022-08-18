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

package grpc

import (
	"context"
	"fmt"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

type boardHandler struct {
	*service
}

// BoardList executes the 'arduino-cli board list' command
// and returns its result.
func (b boardHandler) BoardList(ctx context.Context) ([]*rpc.DetectedPort, error) {
	boardListResp, err := b.serviceClient.BoardList(context.Background(),
		&rpc.BoardListRequest{Instance: b.instance})

	if err != nil {
		err = fmt.Errorf("%s: %w", "Board list error", err)
		return nil, err
	}

	return boardListResp.GetPorts(), nil
}
