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

package dashboard

import (
	"context"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/dashboard"
	"github.com/arduino/arduino-cloud-cli/config"
	v3 "github.com/arduino/iot-client-go/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type showFlags struct {
	deviceId string
}

func initShowCommand() *cobra.Command {
	flags := &showFlags{}
	showCommand := &cobra.Command{
		Use:   "show",
		Short: "Show dashboard",
		Long:  "Show dashboard on Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runShowCommand(flags); err != nil {
				feedback.Errorf("Error during dashboard show: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	showCommand.Flags().StringVarP(&flags.deviceId, "dashboard-id", "d", "", "dashboard ID")

	showCommand.MarkFlagRequired("dashboard-id")

	return showCommand
}

func runShowCommand(flags *showFlags) error {
	logrus.Info("Show dashboard")

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	dash, err := dashboard.Show(context.TODO(), cred, flags.deviceId)
	if err != nil {
		return err
	}

	feedback.PrintResult(showResult{dash})
	return nil
}

type showResult struct {
	dashboard *v3.ArduinoDashboardv3
}

func (r showResult) Data() interface{} {
	return r.dashboard
}

func (r showResult) String() string {
	t, err := yaml.Marshal(r.dashboard)
	if err != nil {
		feedback.Errorf("Error during template parsing: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
	return string(t)
}
