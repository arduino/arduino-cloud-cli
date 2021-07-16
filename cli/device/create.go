package device

import (
	"errors"
	"fmt"
	"strings"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/bcmi-labs/iot-cloud-cli/arduino/grpc"
	"github.com/bcmi-labs/iot-cloud-cli/command/config"
	"github.com/bcmi-labs/iot-cloud-cli/internal/iot"
	"github.com/spf13/cobra"
)

var createFlags struct {
	port string
	name string
	fqbn string
}

type device struct {
	fqbn   string
	serial string
	dType  string
	port   string
}

func initCreateCommand() *cobra.Command {
	createCommand := &cobra.Command{
		Use:   "create",
		Short: "Create a device",
		Long:  "Create a device for Arduino IoT Cloud",
		RunE:  runCreateCommand,
	}
	createCommand.Flags().StringVarP(&createFlags.port, "port", "p", "", "Device port")
	createCommand.Flags().StringVarP(&createFlags.name, "name", "n", "", "Device name")
	createCommand.Flags().StringVarP(&createFlags.fqbn, "fqbn", "b", "", "Device fqbn")
	createCommand.MarkFlagRequired("name")
	return createCommand
}

func runCreateCommand(cmd *cobra.Command, args []string) error {
	fmt.Printf("Creating device with name %s\n", createFlags.name)

	rpcComm, rpcClose, err := grpc.NewClient()
	if err != nil {
		return err
	}
	defer rpcClose()

	ports, err := rpcComm.BoardList()
	if err != nil {
		return err
	}
	dev := deviceFromPorts(ports)
	if dev == nil {
		err = errors.New("no device found")
		return err
	}

	conf, _ := config.Retrieve()
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return err
	}

	fmt.Println("Creating a new device on the cloud")
	devID, err := iotClient.AddDevice(dev.fqbn, createFlags.name, dev.serial, dev.dType)
	if err != nil {
		return err
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
		return err
	}

	fmt.Printf("IoT Cloud device created with ID: %s\n", devID)
	return nil
}

func deviceFromPorts(ports []*rpc.DetectedPort) *device {
	for _, port := range ports {
		if portFilter(port) {
			continue
		}
		board := boardFilter(port.Boards)
		if board != nil {
			t := strings.Split(board.Fqbn, ":")[2]
			dev := &device{board.Fqbn, port.SerialNumber, t, port.Address}
			return dev
		}
	}

	return nil
}

// true -> skip the port
// false -> keep the port
func portFilter(port *rpc.DetectedPort) bool {
	if len(port.Boards) == 0 {
		return true
	}
	if createFlags.port != "" && createFlags.port != port.Address {
		return true
	}
	return false
}

func boardFilter(boards []*rpc.BoardListItem) (board *rpc.BoardListItem) {
	if createFlags.fqbn == "" {
		return boards[0]
	}
	for _, b := range boards {
		if b.Fqbn == createFlags.fqbn {
			return b
		}
	}
	return
}
