package cli

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/arduino/arduino-cli/cli/instance"
	"github.com/arduino/arduino-cli/commands/board"
	"github.com/arduino/arduino-cli/commands/upload"
	"github.com/arduino/arduino-cli/configuration"
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/iot-cloud-cli/arduino"
	"github.com/sirupsen/logrus"
)

type commander struct {
	*rpc.Instance
}

// NewCommander instantiates and returns a new arduino-cli commander that allows to
// programmatically call arduino-cli commands.
// It directly imports the golang packages of the arduino-cli.
func NewCommander() (arduino.Commander, error) {
	// Discard arduino-cli log messages
	logrus.SetOutput(ioutil.Discard)
	// Initialize arduino-cli configuration
	configuration.Settings = configuration.Init(configuration.FindConfigFileInArgsOrWorkingDirectory(os.Args))
	// Create arduino-cli instance, needed to execute arduino-cli commands
	inst, err := instance.CreateInstance()
	if err != nil {
		err = fmt.Errorf("%s: %w", "creating arduino-cli instance", err)
		return nil, err
	}

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

	if _, err := upload.Upload(context.Background(), req, os.Stdout, os.Stderr); err != nil {
		err = fmt.Errorf("%s: %w", "uploading binary", err)
		return err
	}

	return nil
}
