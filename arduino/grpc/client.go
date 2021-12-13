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

package grpc

import (
	"context"
	"fmt"
	"io"
	"time"

	rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/arduino/arduino-cli/rpc/cc/arduino/cli/settings/v1"
	"github.com/arduino/arduino-cloud-cli/arduino"
	"google.golang.org/grpc"
)

type service struct {
	serviceClient  rpc.ArduinoCoreServiceClient
	settingsClient settings.SettingsServiceClient
	instance       *rpc.Instance
}

type client struct {
	boardHandler
	compileHandler
}

// NewClient instantiates and returns a new grpc client that allows to
// programmatically call arduino-cli commands.
// It exploits the grpc interface of the arduino-cli.
// It returns: the client instance, a callback to close the client and an error.
func NewClient() (arduino.Commander, func() error, error) {
	// Establish a connection with the gRPC server, started with the command:
	// arduino-cli daemon
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Second))
	if err != nil {
		err = fmt.Errorf("%s: %w", "cannot connect to arduino-cli rpc server, you can start it by running `arduino-cli daemon`", err)
		return nil, func() error { return nil }, err
	}

	serv := &service{}
	// Create an instance of the gRPC clients.
	serv.serviceClient = rpc.NewArduinoCoreServiceClient(conn)
	serv.settingsClient = settings.NewSettingsServiceClient(conn)
	serv.instance, err = initInstance(serv.serviceClient)
	if err != nil {
		conn.Close()
		err = fmt.Errorf("%s: %w", "creating arduino-cli instance", err)
		return nil, func() error { return nil }, err
	}

	cl := &client{}
	cl.boardHandler = boardHandler{serv}
	cl.compileHandler = compileHandler{serv}

	return cl, conn.Close, nil
}

func initInstance(client rpc.ArduinoCoreServiceClient) (*rpc.Instance, error) {
	initRespStream, err := client.Init(context.Background(), &rpc.InitRequest{})
	if err != nil {
		err = fmt.Errorf("%s: %w", "Error creating server instance", err)
		return nil, err
	}

	var instance *rpc.Instance
	// Loop and consume the server stream until all the setup procedures are done.
	for {
		initResp, err := initRespStream.Recv()
		// The server is done.
		if err == io.EOF {
			break
		}

		// There was an error.
		if err != nil {
			err = fmt.Errorf("%s: %w", "init error", err)
			return nil, err
		}

		// The server sent us a valid instance, let's print its ID.
		if initResp.GetInstance() != nil {
			instance = initResp.GetInstance()
			//fmt.Printf("Got a new instance with ID: %v", instance.GetId())
		}

		// When a download is ongoing, log the progress
		if initResp.GetDownloadProgress() != nil {
			fmt.Printf("DOWNLOAD: %s", initResp.GetDownloadProgress())
		}

		// When an overall task is ongoing, log the progress
		if initResp.GetTaskProgress() != nil {
			fmt.Printf("TASK: %s", initResp.GetTaskProgress())
		}
	}

	return instance, nil
}
