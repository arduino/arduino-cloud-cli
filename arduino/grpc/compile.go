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
	"io"
	"path/filepath"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

type compileHandler struct {
	*service
}

// Compile executes the 'arduino-cli compile' command
// and returns its result.
func (c compileHandler) Compile() error {
	return nil
}

// Upload executes the 'arduino-cli upload -i' command
// and returns its result.
func (c compileHandler) UploadBin(ctx context.Context, fqbn, bin, address, protocol string) error {
	stream, err := c.serviceClient.Upload(context.Background(),
		&rpc.UploadRequest{
			Instance:   c.instance,
			Fqbn:       fqbn,
			SketchPath: filepath.Dir(bin),
			ImportFile: bin,
			Port:       &rpc.Port{Address: address, Protocol: protocol},
			Verbose:    true,
		})

	if err != nil {
		err = fmt.Errorf("%s: %w", "uploading", err)
		return err
	}

	// Wait for the upload to complete
	for {
		_, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			err = fmt.Errorf("%s: %w", "errors during upload", err)
			return err
		}
	}

	return nil
}
