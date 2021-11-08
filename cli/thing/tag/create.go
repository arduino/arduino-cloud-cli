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

package tag

import (
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/tag"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var createTagsFlags struct {
	id   string
	tags map[string]string
}

func InitCreateTagsCommand() *cobra.Command {
	createTagsCommand := &cobra.Command{
		Use:   "create-tags",
		Short: "Create or overwrite tags on a thing",
		Long:  "Create or overwrite tags on a thing of Arduino IoT Cloud",
		Run:   runCreateTagsCommand,
	}
	createTagsCommand.Flags().StringVarP(&createTagsFlags.id, "id", "i", "", "Thing ID")
	createTagsCommand.Flags().StringToStringVar(
		&createTagsFlags.tags,
		"tags",
		nil,
		"Comma-separated list of tags with format <key>=<value>.",
	)
	createTagsCommand.MarkFlagRequired("id")
	createTagsCommand.MarkFlagRequired("tags")
	return createTagsCommand
}

func runCreateTagsCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Creating tags on thing %s\n", createTagsFlags.id)

	params := &tag.CreateTagsParams{
		ID:       createTagsFlags.id,
		Tags:     createTagsFlags.tags,
		Resource: tag.Thing,
	}

	err := tag.CreateTags(params)
	if err != nil {
		feedback.Errorf("Error during thing create-tags: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Info("Tags successfully created")
}
