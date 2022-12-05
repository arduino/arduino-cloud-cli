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

package arduino

import (
	"context"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// Commander of arduino package allows to call
// the arduino-cli commands in a programmatic way.
type Commander interface {
	BoardList(ctx context.Context) ([]*rpc.DetectedPort, error)
	UploadBin(ctx context.Context, fqbn, bin, address, protocol string) error
	//Compile() error
}
