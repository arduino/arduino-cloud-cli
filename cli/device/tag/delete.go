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

type deleteTagsFlags struct {
	id   string
	ids  string
	keys []string
}

func InitDeleteTagsCommand() *cobra.Command {
	flags := &deleteTagsFlags{}
	deleteTagsCommand := &cobra.Command{
		Use:   "delete-tags",
		Short: "Delete tags of a device",
		Long:  "Delete tags of a device of Arduino IoT Cloud",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runDeleteTagsCommand(flags); err != nil {
				feedback.Errorf("Error during device delete-tags: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}
	deleteTagsCommand.Flags().StringVarP(&flags.id, "id", "i", "", "Device ID")
	deleteTagsCommand.Flags().StringVarP(&flags.id, "ids", "", "", "Comma-separated list of Device IDs")
	deleteTagsCommand.Flags().StringSliceVarP(&flags.keys, "keys", "k", nil, "Comma-separated list of keys of tags to delete")
	deleteTagsCommand.MarkFlagRequired("keys")
	return deleteTagsCommand
}

func runDeleteTagsCommand(flags *deleteTagsFlags) error {
	if flags.id == "" && flags.ids == "" {
		return fmt.Errorf("missing required flag(s) \"id\" or \"ids\"")
	}

	if flags.id != "" {
		err := deleteTags(flags.id, flags.keys)
		if err != nil {
			return err
		}
	}
	if flags.ids != "" {
		ids := strings.Split(flags.ids, ",")
		for _, id := range ids {
			id = strings.TrimSpace(id)
			err := deleteTags(id, flags.keys)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func deleteTags(id string, keys []string) error {
	logrus.Infof("Deleting tags with keys %s", keys)

	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}

	params := &tag.DeleteTagsParams{
		ID:       id,
		Keys:     keys,
		Resource: tag.Device,
	}

	err = tag.DeleteTags(context.TODO(), params, cred)
	if err != nil {
		return err
	}

	logrus.Info("Tags successfully deleted")
	return nil
}
