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
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type createVariableFlags struct {
	id              string
	name            string
	variableName    string
	varType         string
	permission      string
	updateStrategy  string
	updateParameter float64
}

func initCreateVariableCommand() *cobra.Command {
	flags := &createVariableFlags{}
	createVariableCommand := &cobra.Command{
		Use:   "create-variable",
		Short: "Add a variable to a thing",
		Long:  "Add a variable to an existing thing of Arduino IoT Cloud, keeping its ID and bindings",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runCreateVariableCommand(cmd, flags); err != nil {
				feedback.Errorf("Error during thing create-variable: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	createVariableCommand.Flags().StringVarP(&flags.id, "id", "i", "", "Thing ID")
	createVariableCommand.Flags().StringVarP(&flags.name, "name", "n", "", "Variable name")
	createVariableCommand.Flags().StringVar(&flags.variableName, "variable-name", "", "Sketch variable name, if different from the variable name")
	createVariableCommand.Flags().StringVarP(&flags.varType, "type", "t", "", "Variable type, as in thing templates (e.g. FLOAT, INT, STATUS, CHARSTRING, HOME_SWITCH)")
	createVariableCommand.Flags().StringVarP(&flags.permission, "permission", "p", "READ_WRITE", "Variable permission: READ_ONLY or READ_WRITE")
	createVariableCommand.Flags().StringVar(&flags.updateStrategy, "update-strategy", "ON_CHANGE", "Update strategy: ON_CHANGE or TIMED")
	createVariableCommand.Flags().Float64Var(
		&flags.updateParameter,
		"update-parameter",
		0,
		"Minimum change that triggers an update (ON_CHANGE) or update period in seconds (TIMED)",
	)
	createVariableCommand.MarkFlagRequired("id")
	createVariableCommand.MarkFlagRequired("name")
	createVariableCommand.MarkFlagRequired("type")
	return createVariableCommand
}

func runCreateVariableCommand(cmd *cobra.Command, flags *createVariableFlags) error {
	logrus.Infof("Adding variable %s to thing %s", flags.name, flags.id)

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	params := &thing.CreateVariableParams{
		ThingID:        flags.id,
		Name:           flags.name,
		Type:           flags.varType,
		Permission:     flags.permission,
		UpdateStrategy: flags.updateStrategy,
	}
	if flags.variableName != "" {
		params.VariableName = &flags.variableName
	}
	if cmd.Flags().Changed("update-parameter") {
		params.UpdateParameter = &flags.updateParameter
	}

	variable, err := thing.CreateVariable(context.TODO(), params, cred)
	if err != nil {
		return err
	}

	feedback.PrintResult(createVariableResult{variable})
	return nil
}

type createVariableResult struct {
	variable *thing.VariableInfo
}

func (r createVariableResult) Data() interface{} {
	return r.variable
}

func (r createVariableResult) String() string {
	return fmt.Sprintf(
		"name: %s\nvariable_name: %s\nid: %s\ntype: %s\npermission: %s\nupdate_strategy: %s",
		r.variable.Name,
		r.variable.VariableName,
		r.variable.ID,
		r.variable.Type,
		r.variable.Permission,
		r.variable.UpdateStrategy,
	)
}
