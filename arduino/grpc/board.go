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
func (b boardHandler) BoardList() ([]*rpc.DetectedPort, error) {
	boardListResp, err := b.serviceClient.BoardList(context.Background(),
		&rpc.BoardListRequest{Instance: b.instance})

	if err != nil {
		err = fmt.Errorf("%s: %w", "Board list error", err)
		return nil, err
	}

	return boardListResp.GetPorts(), nil
}
