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

package dashboard

import (
	"context"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/dashboard"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type extractFlags struct {
	id string
}

func initExtractCommand() *cobra.Command {
	flags := &extractFlags{}
	extractCommand := &cobra.Command{
		Use:   "extract",
		Short: "Extract a template from a dashboard",
		Long:  "Extract a template from a Arduino IoT Cloud dashboard",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runExtractCommand(flags); err != nil {
				feedback.Errorf("Error during dashboard extract: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	extractCommand.Flags().StringVarP(&flags.id, "id", "i", "", "Dashboard ID")
	extractCommand.MarkFlagRequired("id")
	return extractCommand
}

func runExtractCommand(flags *extractFlags) error {
	logrus.Infof("Extracting template from dashboard %s", flags.id)

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	params := &dashboard.ExtractParams{
		ID: flags.id,
	}

	template, err := dashboard.Extract(context.TODO(), params, cred)
	if err != nil {
		return err
	}

	feedback.PrintResult(extractResult{template})
	return nil
}

type extractResult struct {
	template map[string]interface{}
}

func (r extractResult) Data() interface{} {
	return r.template
}

func (r extractResult) String() string {
	t, err := yaml.Marshal(r.template)
	if err != nil {
		feedback.Errorf("Error during template parsing: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}
	return string(t)
}
