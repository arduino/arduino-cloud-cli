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

package ota

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"text/tabwriter"

	"gotest.tools/assert"
)

func TestVersionWithCompressionEnabled(t *testing.T) {

	version := Version{
		Compression: true,
	}

	expected := []byte{0, 0, 0, 0, 0, 0, 0, 0x40}
	actual := version.Bytes()

	// create a tabwriter for formatting the output
	w := new(tabwriter.Writer)

	// Format in tab-separated columns with a tab stop of 8.
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)

	fmt.Fprintf(w, "Binary:\t%0.8bb (expected)\n", expected)
	fmt.Fprintf(w, "Binary:\t%0.8bb (actual)\n", actual)
	w.Flush()

	res := bytes.Compare(expected, actual)
	assert.Assert(t, res == 0) // 0 means equal
}
