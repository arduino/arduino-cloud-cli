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
	"strconv"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/arduino-cloud-cli/command/thing"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type listVariablesFlags struct {
	id string
}

func initListVariablesCommand() *cobra.Command {
	flags := &listVariablesFlags{}
	listVariablesCommand := &cobra.Command{
		Use:   "list-variables",
		Short: "List the variables of a thing",
		Long:  "List the variables of a thing of Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runListVariablesCommand(flags); err != nil {
				feedback.Errorf("Error during thing list-variables: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	listVariablesCommand.Flags().StringVarP(&flags.id, "id", "i", "", "Thing ID")
	listVariablesCommand.MarkFlagRequired("id")
	return listVariablesCommand
}

func runListVariablesCommand(flags *listVariablesFlags) error {
	logrus.Infof("Listing variables of thing %s", flags.id)

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	variables, err := thing.ListVariables(context.TODO(), flags.id, cred)
	if err != nil {
		return err
	}

	feedback.PrintResult(listVariablesResult{variables})
	return nil
}

type listVariablesResult struct {
	variables []thing.VariableInfo
}

func (r listVariablesResult) Data() interface{} {
	return r.variables
}

func (r listVariablesResult) String() string {
	if len(r.variables) == 0 {
		return "No variables found."
	}
	t := table.New()
	t.SetHeader("Name", "Variable name", "ID", "Type", "Permission", "Update strategy", "Update parameter")
	for _, v := range r.variables {
		updateParameter := ""
		if v.UpdateParameter != nil {
			updateParameter = strconv.FormatFloat(*v.UpdateParameter, 'f', -1, 64)
		}
		t.AddRow(v.Name, v.VariableName, v.ID, v.Type, v.Permission, v.UpdateStrategy, updateParameter)
	}
	return t.Render()
}
