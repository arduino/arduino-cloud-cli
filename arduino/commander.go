package arduino

import (
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// Commander of arduino package allows to call
// the arduino-cli commands in a programmatic way
type Commander interface {
	BoardList() ([]*rpc.DetectedPort, error)
	UploadBin(fqbn, path, port string) error
	Compile() error
}
