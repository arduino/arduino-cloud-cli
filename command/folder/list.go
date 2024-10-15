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

package folder

import (
	"context"

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
)

func ListFolders(ctx context.Context, cred *config.Credentials) error {
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return err
	}

	fold, err := iotClient.FoldersList(ctx)
	if err != nil {
		return err
	}

	if fold == nil {
		feedback.Print("No folders found")
	} else {
		folders := &Folders{}
		for _, f := range fold {
			trFolder := Folder{
				ID:        f.Id,
				Name:      f.Name,
				CreatedAt: f.CreatedAt,
				UpdatedAt: f.UpdatedAt,
			}
			if f.Parent != nil {
				trFolder.Parent = &PathNode{
					FolderId:   f.Parent.FolderId,
					FolderName: f.Parent.FolderName,
				}
			}
			if f.Path != nil {
				trFolder.Path = make([]PathNode, 0, len(f.Path))
				for _, p := range f.Path {
					trFolder.Path = append(trFolder.Path, PathNode{
						FolderId:   p.FolderId,
						FolderName: p.FolderName,
					})
				}
			}
			folders.Folders = append(folders.Folders, trFolder)
		}
		feedback.PrintResult(folders)
	}

	return nil
}
