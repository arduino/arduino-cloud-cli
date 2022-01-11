package lzss

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name    string
		infile  string
		outfile string
	}{
		{
			name:    "blink",
			infile:  "testdata/blink.bin",
			outfile: "testdata/blink.lzss",
		},
		{
			name:    "cloud sketch",
			infile:  "testdata/cloud.bin",
			outfile: "testdata/cloud.lzss",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := ioutil.ReadFile(tt.infile)
			if err != nil {
				t.Fatal("couldn't open test file")
			}

			want, err := ioutil.ReadFile(tt.outfile)
			if err != nil {
				t.Fatal("couldn't open test file")
			}

			got := Encode(input)
			if !bytes.Equal(want, got) {
				t.Error("encoding failed")
			}
		})
	}
}
