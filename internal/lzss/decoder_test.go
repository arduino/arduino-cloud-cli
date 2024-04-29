package lzss

import (
	"bytes"
	"os"
	"testing"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		infile  string
		outfile string
	}{
		{
			name:    "blink",
			infile:  "testdata/blink.lzss",
			outfile: "testdata/blink.bin",
		},
		{
			name:    "cloud sketch",
			infile:  "testdata/cloud.lzss",
			outfile: "testdata/cloud.bin",
		},
		{
			name:    "empty binary",
			infile:  "testdata/empty.lzss",
			outfile: "testdata/empty.bin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, err := os.ReadFile(tt.infile)
			if err != nil {
				t.Fatal("couldn't open test file")
			}

			want, err := os.ReadFile(tt.outfile)
			if err != nil {
				t.Fatal("couldn't open test file")
			}

			got := Decompress(input)
			if !bytes.Equal(want, got) {
				t.Error("decoding failed", want, got)
			}
		})
	}
}
