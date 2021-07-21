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
func (c compileHandler) UploadBin(fqbn, bin, port string) error {
	stream, err := c.serviceClient.Upload(context.Background(),
		&rpc.UploadRequest{
			Instance:   c.instance,
			Fqbn:       fqbn,
			SketchPath: filepath.Dir(bin),
			ImportFile: bin,
			Port:       port,
			Verbose:    true,
		})

	if err != nil {
		err = fmt.Errorf("%s: %w", "uploading", err)
		return err
	}

	// Wait for the upload to complete
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			err = fmt.Errorf("%s: %w", "errors during upload", err)
			return err
		}
		if resp.ErrStream != nil {
			err = fmt.Errorf("%s: %w", "errors during upload", err)
			return err
		}
	}

	return nil
}
