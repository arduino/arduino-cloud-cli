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

package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplateLIsting(t *testing.T) {
	tmpDir := t.TempDir()
	path := ".testdata/template_ok-rp-with-binaries.tino"
	file, err := extractBinary(&path, tmpDir)
	assert.Nil(t, err)
	assert.NotNil(t, file)
}
