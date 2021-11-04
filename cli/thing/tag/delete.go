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

var deleteTagsFlags struct {
	id   string
	keys []string
}

func InitDeleteTagsCommand() *cobra.Command {
	deleteTagsCommand := &cobra.Command{
		Use:   "delete-tags",
		Short: "Delete tags of a thing",
		Long:  "Delete tags of a thing of Arduino IoT Cloud",
		Run:   runDeleteTagsCommand,
	}

	deleteTagsCommand.Flags().StringVarP(&deleteTagsFlags.id, "id", "i", "", "Thing ID")
	deleteTagsCommand.Flags().StringSliceVarP(&deleteTagsFlags.keys, "keys", "k", nil, "List of comma-separated tag keys to delete")

	deleteTagsCommand.MarkFlagRequired("id")
	deleteTagsCommand.MarkFlagRequired("keys")
	return deleteTagsCommand
}

func runDeleteTagsCommand(cmd *cobra.Command, args []string) {
	logrus.Infof("Deleting tags %s\n", deleteTagsFlags.keys)

	params := &tag.DeleteTagsParams{
		ID:       deleteTagsFlags.id,
		Keys:     deleteTagsFlags.keys,
		Resource: tag.Thing,
	}

	err := tag.DeleteTags(params)
	if err != nil {
		feedback.Errorf("Error during thing delete-tags: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	logrus.Info("Tags successfully deleted")
}
