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
	"fmt"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/thing"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cloneFlags struct {
	name    string
	cloneID string
}

func initCloneCommand() *cobra.Command {
	cloneCommand := &cobra.Command{
		Use:   "clone",
		Short: "Clone a thing",
		Long:  "Clone a thing for Arduino IoT Cloud",
		Run:   runCloneCommand,
	}
	cloneCommand.Flags().StringVarP(&cloneFlags.name, "name", "n", "", "Thing name")
	cloneCommand.Flags().StringVarP(&cloneFlags.cloneID, "clone-id", "c", "", "ID of Thing to be cloned")
	cloneCommand.MarkFlagRequired("name")
	cloneCommand.MarkFlagRequired("clone-id")
	return cloneCommand
}

func runCloneCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Cloning thing %s into a new thing called %s\n", cloneFlags.cloneID, cloneFlags.name)

	params := &thing.CloneParams{
		Name:    cloneFlags.name,
		CloneID: cloneFlags.cloneID,
	}

	thing, err := thing.Clone(params)
	if err != nil {
		feedback.Errorf("Error during thing clone: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.PrintResult(cloneResult{thing})
}

type cloneResult struct {
	thing *thing.ThingInfo
}

func (r cloneResult) Data() interface{} {
	return r.thing
}

func (r cloneResult) String() string {
	return fmt.Sprintf(
		"name: %s\nid: %s\ndevice-id: %s\nvariables: %s",
		r.thing.Name,
		r.thing.ID,
		r.thing.DeviceID,
		strings.Join(r.thing.Variables, ", "),
	)
}
