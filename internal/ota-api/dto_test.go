// This file is part of arduino-cloud-cli.
//
// Copyright ARDUINO SRL http://www.arduino.cc/)
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

package otaapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProgressBar_notCompletePct(t *testing.T) {
	firmwareSize := int64(25665 * 2)
	bar := formatStateData("fetch", "25665", firmwareSize, false)
	assert.Equal(t, "[==========          ] 50% (firmware size: 51330 bytes)", bar)
}

func TestProgressBar_ifFlashState_goTo100Pct(t *testing.T) {
	firmwareSize := int64(25665 * 2)
	bar := formatStateData("fetch", "25665", firmwareSize, true) // If in flash status, go to 100%
	assert.Equal(t, "[====================] 100% (firmware size: 51330 bytes)", bar)
}
