package arduino

import (
	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
)

// Client of arduino package allows to call
// the arduino-cli commands in a programmatic way
type Client interface {
	BoardList() ([]*rpc.DetectedPort, error)
	Compile() error
}
