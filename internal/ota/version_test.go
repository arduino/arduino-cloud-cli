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
	actual := version.AsBytes()

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
