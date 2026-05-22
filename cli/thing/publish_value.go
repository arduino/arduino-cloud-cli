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

// publishParams contains the parameters needed to publish a value to a thing.
type publishParams struct {
	ThingID        string            // ID of the thing
	PropertyName   string            // Name of the property
	Value          string            // Value to publish
	MultipleValues map[string]string // Map of property names and values to publish
}

func initPublishCommand() *cobra.Command {
	flags := &publishParams{}
	publishCommand := &cobra.Command{
		Use:   "publish",
		Short: "Publish a value to a variable of a thing",
		Long:  "Publish a value to a variable of a thing for Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runPublishValueCommand(flags); err != nil {
				feedback.Errorf("Error during thing publish: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	publishCommand.Flags().StringVarP(&flags.ThingID, "thing-id", "t", "", "Thing ID")
	publishCommand.Flags().StringVarP(&flags.PropertyName, "property", "p", "", "Property name")
	publishCommand.Flags().StringVarP(&flags.Value, "value", "s", "", "Value to publish")
	publishCommand.Flags().StringToStringVarP(&flags.MultipleValues, "multiple-values", "m", nil,
		"Map of property names and values to publish. Ex: 'coloredprop={\"bri\": 100.0, \"hue\": 203.0, \"sat\": 100.0, \"swi\": true},simpleprop=hello'")
	publishCommand.MarkFlagRequired("thing-id")

	return publishCommand
}

func runPublishValueCommand(flags *publishParams) error {
	logrus.Infof("Publishing value to thing %s", flags.ThingID)

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	if flags.MultipleValues != nil {
		err = runPublishMultipleValuesCommand(flags, cred)
	} else {
		err = runPublishSingleValueCommand(flags, cred)
	}

	if err != nil {
		return err
	}

	feedback.Print("Variable value updated successfully")
	return nil
}

func runPublishMultipleValuesCommand(flags *publishParams, cred *config.Credentials) error {
	logrus.Infof("Publishing multiple values to thing %s", flags.ThingID)
	return thing.MultiplePublish(context.TODO(), flags.ThingID, flags.MultipleValues, cred)
}

func runPublishSingleValueCommand(flags *publishParams, cred *config.Credentials) error {
	if flags.PropertyName == "" {
		return fmt.Errorf("property name is required when publishing a single value")
	}
	if flags.Value == "" {
		return fmt.Errorf("value is required when publishing a single value")
	}

	logrus.Infof("Publishing value to thing %s, property %s", flags.ThingID, flags.PropertyName)
	return thing.MultiplePublish(context.TODO(), flags.ThingID, map[string]string{flags.PropertyName: flags.Value}, cred)
}
