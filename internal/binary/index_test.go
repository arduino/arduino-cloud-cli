package binary

import (
	"testing"
)

func TestFindProvisionBin(t *testing.T) {
	var (
		fqbnOK1      = "arduino:samd:nano_33_iot"
		fqbnOK2      = "arduino:samd:mkrwifi1010"
		fqbnNotFound = "arduino:mbed_nano:nano33ble"
	)
	index := &Index{
		Boards: []IndexBoard{
			{Fqbn: fqbnOK1, Provision: &IndexBin{URL: "mkr"}},
			{Fqbn: fqbnOK2, Provision: &IndexBin{URL: "nano"}},
		},
	}

	bin := index.FindProvisionBin(fqbnOK2)
	if bin == nil {
		t.Fatal("provision binary not found")
	}

	bin = index.FindProvisionBin(fqbnNotFound)
	if bin != nil {
		t.Fatalf("provision binary should've not be found, but got: %v", bin)
	}
}

func TestLoadIndex(t *testing.T) {
	_, err := LoadIndex()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
