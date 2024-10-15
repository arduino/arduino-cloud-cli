// This file is part of arduino-cloud-cli.
//
// Copyright (C) 2024 ARDUINO SA (http://www.arduino.cc/)
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

package folders

import (
	"context"
	"fmt"
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/folder"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/spf13/cobra"
)

func initFoldersListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List folders",
		Long:  "List available folders",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runFoldersListCommand(); err != nil {
				feedback.Errorf("Error during folders list: %v", err)
				os.Exit(errorcodes.ErrGeneric)
			}
		},
	}

	return listCommand
}

func runFoldersListCommand() error {
	cred, err := config.RetrieveCredentials()
	if err != nil {
		return fmt.Errorf("retrieving credentials: %w", err)
	}
	ctx := context.Background()
	return folder.ListFolders(ctx, cred)
}
