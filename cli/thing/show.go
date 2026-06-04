// This file is part of arduino-cloud-cli.
//
// Copyright (C) ARDUINO SRL (http://www.arduino.cc)
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

package thing

import (
	"context"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/thing"
	"github.com/arduino/arduino-cloud-cli/config"
	v3 "github.com/arduino/iot-client-go/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type showFlags struct {
	thingId string
}

func initShowCommand() *cobra.Command {
	flags := &showFlags{}
	showCommand := &cobra.Command{
		Use:   "show",
		Short: "Show thing",
		Long:  "Show thing on Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runShowCommand(flags); err != nil {
				feedback.Errorf("Error during thing show: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	showCommand.Flags().StringVarP(&flags.thingId, "thing-id", "t", "", "thing ID")

	showCommand.MarkFlagRequired("thing-id")

	return showCommand
}

func runShowCommand(flags *showFlags) error {
	logrus.Info("Show thing")

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	thing, err := thing.Show(context.TODO(), cred, flags.thingId)
	if err != nil {
		return err
	}

	feedback.PrintResult(showResult{thing})
	return nil
}

type showResult struct {
	thing *v3.ArduinoThing
}

func (r showResult) Data() interface{} {
	return r.thing
}

func (r showResult) String() string {
	t, err := yaml.Marshal(r.thing)
	if err != nil {
		feedback.Errorf("Error during template parsing: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
	return string(t)
}
