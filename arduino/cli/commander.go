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

package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/commands/upload"
	"github.com/arduino/arduino-cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cloud-cli/arduino"
	"github.com/sirupsen/logrus"
)

type commander struct {
	*rpc.Instance
}

// NewCommander instantiates and returns a new arduino-cli commander that allows to
// programmatically call arduino-cli commands.
// It directly imports the golang packages of the arduino-cli.
func NewCommander() (arduino.Commander, error) {
	// Discard arduino-cli log info messages.
	logrus.SetLevel(logrus.PanicLevel)

	// Initialize arduino-cli configuration.
	configuration.Settings = configuration.Init(configuration.FindConfigFileInArgsOrWorkingDirectory(os.Args))

	// Create and init an arduino-cli instance, needed to execute arduino-cli commands.
	inst, err := instance.Create()
	if err != nil {
		err = fmt.Errorf("creating arduino-cli instance: %w", err)
		return nil, err
	}
	errs := instance.Init(inst)
	if len(errs) > 0 {
		err = errors.New("initializing arduino-cli instance: received errors: ")
		for _, e := range errs {
			err = fmt.Errorf("%v: %v; ", err, e)
		}
		return nil, err
	}

	// Re-enable info level log
	logrus.SetLevel(logrus.InfoLevel)
	cmd := &commander{inst}
	return cmd, nil
}

// BoardList executes the 'arduino-cli board list' command
// and returns its result.
func (c *commander) BoardList(ctx context.Context) ([]*rpc.DetectedPort, error) {
	req := &rpc.BoardListRequest{
		Instance: c.Instance,
		Timeout:  time.Second.Milliseconds(),
	}

	// There is no obvious way to cancel the execution of this command.
	// So, we execute it in a goroutine and leave it running alone if ctx gets cancelled.
	type resp struct {
		err   error
		ports []*rpc.DetectedPort
	}
	quit := make(chan resp, 1)
	go func() {
		ports, err := board.List(req)
		quit <- resp{err: err, ports: ports}
		close(quit)
	}()

	// Wait for the command to complete or the context to be terminated.
	select {
	case <-ctx.Done():
		return nil, errors.New("board list command cancelled")
	case r := <-quit:
		if r.err != nil {
			return nil, fmt.Errorf("executing board list command: %w", r.err)
		}
		return r.ports, nil
	}
}

// UploadBin executes the 'arduino-cli upload -i' command
// and returns its result.
func (c *commander) UploadBin(ctx context.Context, fqbn, bin, address, protocol string) error {
	req := &rpc.UploadRequest{
		Instance:   c.Instance,
		Fqbn:       fqbn,
		SketchPath: filepath.Dir(bin),
		ImportFile: bin,
		Port:       &rpc.Port{Address: address, Protocol: protocol},
		Verbose:    false,
	}
	l := logrus.StandardLogger().WithField("source", "arduino-cli").Writer()

	// There is no obvious way to cancel the execution of this command.
	// So, we execute it in a goroutine and leave it running if ctx gets cancelled.
	quit := make(chan error, 1)
	go func() {
		_, err := upload.Upload(ctx, req, l, l)
		quit <- err
		close(quit)
	}()

	// Wait for the upload to complete or the context to be terminated.
	select {
	case <-ctx.Done():
		return errors.New("upload cancelled")
	case err := <-quit:
		if err != nil {
			return fmt.Errorf("uploading binary: %w", err)
		}
		return nil
	}
}
