package device

import (
	"context"
	"errors"
	"strings"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cloud-cli/arduino/cli"
	"go.bug.st/cleanup"
)

func ListAllConnectedBoardsWithCrypto(fqbn *string) ([]*Board, error) {
	comm, err := ListAllConnectedBoards()
	if err != nil {
		return nil, err
	}
	var withcrypto []*Board
	for _, b := range comm {
		if len(*fqbn) > 0 && b.Fqbn != *fqbn {
			// Skipp not matching board
			continue
		}
		if b.isCrypto() {
			withcrypto = append(withcrypto, b)
		}
	}
	return withcrypto, nil
}

func ListAllConnectedBoards() ([]*Board, error) {
	comm, err := cli.NewCommander()
	if err != nil {
		return nil, err
	}

	ctx, cancel := cleanup.InterruptableContext(context.Background())
	defer cancel()

	ports, err := comm.BoardList(ctx)
	if err != nil {
		return nil, err
	}

	board := boardsFromPorts(ports, nil)
	if board == nil {
		err = errors.New("no board found")
		return nil, err
	}

	return board, nil
}

// boardsFromPorts returns boards that matches all the criteria
func boardsFromPorts(ports []*rpc.DetectedPort, fqbn *string) []*Board {
	var boards []*Board
	for _, port := range ports {
		boardsFound := boardsFilter(port.MatchingBoards, fqbn)
		if len(boardsFound) > 0 {
			for _, boardFound := range boardsFound {
				b := &Board{
					Fqbn:     boardFound.Fqbn,
					Serial:   port.Port.Properties["serialNumber"],
					DType:    strings.Split(boardFound.Fqbn, ":")[2],
					Address:  port.Port.Address,
					Protocol: port.Port.Protocol,
				}
				b.isCrypto()
				boards = append(boards, b)
			}
		}
	}
	return boards
}

func boardsFilter(boards []*rpc.BoardListItem, fqbn *string) (board []*rpc.BoardListItem) {
	if fqbn == nil {
		return boards
	}
	var filtered []*rpc.BoardListItem
	for _, b := range boards {
		if b.Fqbn == *fqbn {
			filtered = append(filtered, b)
		}
	}
	return filtered
}
