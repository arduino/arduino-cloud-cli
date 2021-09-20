package ota

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"

	inota "github.com/arduino/arduino-cloud-cli/internal/ota"
)

var (
	arduinoVendorID = "2341"
	fqbnToPID       = map[string]string{
		"arduino:samd:nano_33_iot":            "8057",
		"arduino:samd:mkr1000":                "804E",
		"arduino:samd:mkrgsm1400":             "8052",
		"arduino:samd:mkrnb1500":              "8055",
		"arduino:samd:mkrwifi1010":            "8054",
		"arduino:mbed_nano:nanorp2040connect": "005E",
		"arduino:mbed_portenta:envie_m7":      "025B",
	}
)

// Generate takes a .bin file and generates a .ota file.
func Generate(binFile string, outFile string, fqbn string) error {
	productID, ok := fqbnToPID[fqbn]
	if !ok {
		return errors.New("fqbn not valid")
	}

	data, err := ioutil.ReadFile(binFile)
	if err != nil {
		return err
	}

	var w bytes.Buffer
	otaWriter := inota.NewWriter(&w, arduinoVendorID, productID)
	_, err = otaWriter.Write(data)
	if err != nil {
		return err
	}
	otaWriter.Close()

	err = ioutil.WriteFile(outFile, w.Bytes(), os.FileMode(0644))
	if err != nil {
		return err
	}

	return nil
}
