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

package thing

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/thing"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type cloneFlags struct {
	name    string
	cloneID string
}

func initCloneCommand() *cobra.Command {
	flags := &cloneFlags{}
	cloneCommand := &cobra.Command{
		Use:   "clone",
		Short: "Clone a thing",
		Long:  "Clone a thing for Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runCloneCommand(flags); err != nil {
				feedback.Errorf("Error during thing clone: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	cloneCommand.Flags().StringVarP(&flags.name, "name", "n", "", "Thing name")
	cloneCommand.Flags().StringVarP(&flags.cloneID, "clone-id", "c", "", "ID of Thing to be cloned")
	cloneCommand.MarkFlagRequired("name")
	cloneCommand.MarkFlagRequired("clone-id")
	return cloneCommand
}

func runCloneCommand(flags *cloneFlags) error {
	logrus.Infof("Cloning thing %s into a new thing called %s", flags.cloneID, flags.name)

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	params := &thing.CloneParams{
		Name:    flags.name,
		CloneID: flags.cloneID,
	}

	thing, err := thing.Clone(context.TODO(), params, cred)
	if err != nil {
		return err
	}

	feedback.PrintResult(cloneResult{thing})
	return nil
}

type cloneResult struct {
	thing *thing.ThingInfo
}

func (r cloneResult) Data() interface{} {
	return r.thing
}

func (r cloneResult) String() string {
	return fmt.Sprintf(
		"name: %s\nid: %s\ndevice_id: %s\nvariables: %s",
		r.thing.Name,
		r.thing.ID,
		r.thing.DeviceID,
		strings.Join(r.thing.Variables, ", "),
	)
}
