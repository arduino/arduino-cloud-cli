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

type deleteVariableFlags struct {
	id         string
	variableID string
}

func initDeleteVariableCommand() *cobra.Command {
	flags := &deleteVariableFlags{}
	deleteVariableCommand := &cobra.Command{
		Use:   "delete-variable",
		Short: "Delete a variable of a thing",
		Long:  "Delete a variable of a thing of Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runDeleteVariableCommand(flags); err != nil {
				feedback.Errorf("Error during thing delete-variable: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	deleteVariableCommand.Flags().StringVarP(&flags.id, "id", "i", "", "Thing ID")
	deleteVariableCommand.Flags().StringVar(&flags.variableID, "variable-id", "", "ID of the variable to delete, as printed by 'thing list-variables'")
	deleteVariableCommand.MarkFlagRequired("id")
	deleteVariableCommand.MarkFlagRequired("variable-id")
	return deleteVariableCommand
}

func runDeleteVariableCommand(flags *deleteVariableFlags) error {
	logrus.Infof("Deleting variable %s of thing %s", flags.variableID, flags.id)

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	err = thing.DeleteVariable(context.TODO(), flags.id, flags.variableID, cred)
	if err != nil {
		return err
	}

	logrus.Info("Variable successfully deleted")
	return nil
}
