package device

import (
	"errors"
	"fmt"
	"strings"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/iot-cloud-cli/arduino/grpc"
	"github.com/arduino/iot-cloud-cli/internal/config"
	"github.com/arduino/iot-cloud-cli/internal/iot"
)

// CreateParams contains the paramters needed
// to find the device to be provisioned.
// If Port is an empty string, then each serial port is analyzed.
// If Fqbn is an empty string, then the first device found gets selected.
type CreateParams struct {
	Port string
	Name string
	Fqbn string
}

type device struct {
	fqbn   string
	serial string
	dType  string
	port   string
}

// Create command is used to provision a new arduino device
// and to add it to the arduino iot cloud.
func Create(params *CreateParams) (string, error) {
	rpcComm, rpcClose, err := grpc.NewClient()
	if err != nil {
		return "", err
	}
	defer rpcClose()

	ports, err := rpcComm.BoardList()
	if err != nil {
		return "", err
	}
	dev := deviceFromPorts(ports, params)
	if dev == nil {
		err = errors.New("no device found")
		return "", err
	}

	conf, err := config.Retrieve()
	if err != nil {
		return "", err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return "", err
	}

	fmt.Println("Creating a new device on the cloud")
	devID, err := iotClient.AddDevice(dev.fqbn, params.Name, dev.serial, dev.dType)
	if err != nil {
		return "", err
	}

	prov := &provision{
		Commander: rpcComm,
		Client:    iotClient,
		dev:       dev,
		id:        devID}
	err = prov.run()
	if err != nil {
		// TODO: delete the device on iot cloud
		err = fmt.Errorf("%s: %w", "cannot provision device", err)
		return "", err
	}

	return devID, nil
}

// deviceFromPorts returns a board that matches all the criteria
// passed in. If no criteria are passed, it returns the first device found.
func deviceFromPorts(ports []*rpc.DetectedPort, params *CreateParams) *device {
	for _, port := range ports {
		if portFilter(port, params) {
			continue
		}
		board := boardFilter(port.Boards, params)
		if board != nil {
			t := strings.Split(board.Fqbn, ":")[2]
			dev := &device{board.Fqbn, port.SerialNumber, t, port.Address}
			return dev
		}
	}

	return nil
}

// portFilter filters out the given port if the port parameter is not an empty string
// and if they do not match.
// It returns:
// true -> to skip the port
// false -> to keep the port
func portFilter(port *rpc.DetectedPort, params *CreateParams) bool {
	if len(port.Boards) == 0 {
		return true
	}
	if params.Port != "" && params.Port != port.Address {
		return true
	}
	return false
}

// boardFilter looks for a board which has the same fqbn passed as parameter.
// It returns:
// - a board if it is found.
// - nil if no board matching the fqbn parameter is found.
func boardFilter(boards []*rpc.BoardListItem, params *CreateParams) (board *rpc.BoardListItem) {
	if params.Fqbn == "" {
		return boards[0]
	}
	for _, b := range boards {
		if b.Fqbn == params.Fqbn {
			return b
		}
	}
	return
}
