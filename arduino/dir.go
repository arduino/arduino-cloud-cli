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

package arduino

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/arduino/go-paths-helper"
	"github.com/arduino/go-win32-utils"
)

// DataDir returns the Arduino default data directory (arduino15).
func DataDir() (*paths.Path, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("unable to get user home dir: %w", err)
	}

	var path *paths.Path
	switch runtime.GOOS {
	case "darwin":
		path = paths.New(filepath.Join(userHomeDir, "Library", "Arduino15"))
	case "windows":
		localAppDataPath, err := win32.GetLocalAppDataFolder()
		if err != nil {
			return nil, fmt.Errorf("unable to get local app data folder: %w", err)
		}
		path = paths.New(filepath.Join(localAppDataPath, "Arduino15"))
	default: // linux, android, *bsd, plan9 and other Unix-like systems
		path = paths.New(filepath.Join(userHomeDir, ".arduino15"))
	}

	return path, nil
}
