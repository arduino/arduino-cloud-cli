// This file is part of arduino-cloud-cli.
//
// Copyright (C) ARDUINO SRL (http://www.arduino.cc/)
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

package tag

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/tag"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type createTagsFlags struct {
	id   string
	ids  string
	tags map[string]string
}

func InitCreateTagsCommand() *cobra.Command {
	flags := &createTagsFlags{}
	createTagsCommand := &cobra.Command{
		Use:   "create-tags",
		Short: "Create or overwrite tags on a device",
		Long:  "Create or overwrite tags on a device of Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runCreateTagsCommand(flags); err != nil {
				feedback.Errorf("Error during device create-tags: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	createTagsCommand.Flags().StringVarP(&flags.id, "id", "i", "", "Device ID")
	createTagsCommand.Flags().StringVarP(&flags.ids, "ids", "", "", "Comma-separated list of Device IDs")
	createTagsCommand.Flags().StringToStringVar(
		&flags.tags,
		"tags",
		nil,
		"Comma-separated list of tags with format <key>=<value>.",
	)
	createTagsCommand.MarkFlagRequired("tags")
	return createTagsCommand
}

func runCreateTagsCommand(flags *createTagsFlags) error {
	if flags.id == "" && flags.ids == "" {
		return fmt.Errorf("missing required flag(s) \"id\" or \"ids\"")
	}

	if flags.id != "" {
		if err := creteTag(flags.id, flags.tags); err != nil {
			return err
		}
	}
	if flags.ids != "" {
		idsArray := strings.Split(flags.ids, ",")
		for _, id := range idsArray {
			id = strings.TrimSpace(id)
			if err := creteTag(id, flags.tags); err != nil {
				return err
			}
		}
	}
	return nil
}

func creteTag(id string, tags map[string]string) error {
	logrus.Infof("Creating tags on device %s", id)

	params := &tag.CreateTagsParams{
		ID:       id,
		Tags:     tags,
		Resource: tag.Device,
	}

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	if err = tag.CreateTags(context.TODO(), params, cred); err != nil {
		return err
	}

	logrus.Info("Tags successfully created")
	return nil
}
