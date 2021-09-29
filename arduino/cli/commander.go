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
	"fmt"
	"os"
	"path/filepath"

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
	// Discard arduino-cli log info messages
	logrus.SetLevel(logrus.PanicLevel)
	// Initialize arduino-cli configuration
	configuration.Settings = configuration.Init(configuration.FindConfigFileInArgsOrWorkingDirectory(os.Args))
	// Create arduino-cli instance, needed to execute arduino-cli commands
	inst, err := instance.CreateInstance()
	if err != nil {
		err = fmt.Errorf("%s: %w", "creating arduino-cli instance", err)
		return nil, err
	}

	// Re-enable info level log
	logrus.SetLevel(logrus.InfoLevel)
	cmd := &commander{inst}
	return cmd, nil
}

// BoardList executes the 'arduino-cli board list' command
// and returns its result.
func (c *commander) BoardList() ([]*rpc.DetectedPort, error) {
	ports, err := board.List(c.GetId())
	if err != nil {
		err = fmt.Errorf("%s: %w", "detecting boards", err)
		return nil, err
	}
	return ports, nil
}

// UploadBin executes the 'arduino-cli upload -i' command
// and returns its result.
func (c *commander) UploadBin(fqbn, bin, port string) error {
	req := &rpc.UploadRequest{
		Instance:   c.Instance,
		Fqbn:       fqbn,
		SketchPath: filepath.Dir(bin),
		ImportFile: bin,
		Port:       port,
		Verbose:    false,
	}

	l := logrus.StandardLogger().WithField("source", "arduino-cli").Writer()
	if _, err := upload.Upload(context.Background(), req, l, l); err != nil {
		err = fmt.Errorf("%s: %w", "uploading binary", err)
		return err
	}
	return nil
}
