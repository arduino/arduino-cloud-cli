package ota

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/arduino/arduino-cloud-cli/internal/iot/mocks"
	iotclient "github.com/arduino/iot-client-go"
	"github.com/stretchr/testify/mock"
)

const testFilename = "testdata/empty.bin"

func TestRun(t *testing.T) {
	var (
		failID1 = "00000000-b39d-47a2-adf3-d26cdf474707"
		failID2 = "00000000-9efd-4670-a478-df76ebdeeb4f"
		okID1   = "11111111-4838-4f46-8930-d735c5b76cd1"
		okID2   = "11111111-003f-42f9-a80c-85a1de36753b"
		okID3   = "11111111-dac4-4a6a-80a4-698062fe2af5"
	)
	mockClient := &mocks.Client{}
	mockDeviceOTA := func(id string, file *os.File, expireMins int) error {
		time.Sleep(100 * time.Millisecond)
		if strings.Split(id, "-")[0] == "00000000" {
			return errors.New("err")
		}
		return nil
	}
	mockClient.On("DeviceOTA", mock.Anything, mock.Anything, mock.Anything).Return(mockDeviceOTA, nil)

	good, fail, err := run(mockClient, []string{okID1, failID1, okID2, failID2, okID3}, testFilename, 0)
	if len(err) != 2 {
		t.Errorf("two errors should have been returned, got %d: %v", len(err), err)
	}
	if len(fail) != 2 {
		t.Errorf("two updates should have failed, got %d: %v", len(fail), fail)
	}
	if len(good) != 3 {
		t.Errorf("two updates should have succeded, got %d: %v", len(good), good)
	}
}

func TestValidateDevices(t *testing.T) {
	var (
		correctFQBN = "arduino:samd:nano_33_iot"
		wrongFQBN   = "arduino:samd:mkrwifi1010"

		idCorrect1 = "88d683a4-525e-423d-bad2-66a54d3585df"
		idCorrect2 = "84b593fa-86dd-4954-904d-60f657158715"
		idNotValid = "e3a3a667-a859-4317-be97-a61fb6f63487"
		idNotFound = "deb17b7f-b39d-47a2-adf3-d26cdf474707"
	)

	mockClient := &mocks.Client{}
	mockDeviceList := func(tags map[string]string) []iotclient.ArduinoDevicev2 {
		return []iotclient.ArduinoDevicev2{
			{Id: idCorrect1, Fqbn: correctFQBN},
			{Id: idCorrect2, Fqbn: correctFQBN},
			{Id: idNotValid, Fqbn: wrongFQBN},
		}
	}
	mockClient.On("DeviceList", mock.Anything).Return(mockDeviceList, nil)

	ids := []string{
		idCorrect1,
		idNotFound,
		idCorrect2,
		idNotValid,
	}
	v, i, d, _ := validateDevices(mockClient, ids, correctFQBN)

	if len(v) != 2 {
		t.Errorf("expected 2 valid devices, but found %d: %v", len(v), v)
	}

	if len(i) != 2 {
		t.Errorf("expected 2 invalid devices, but found %d: %v", len(i), i)
	}

	if len(d) != 2 {
		t.Errorf("expected 2 error details, but found %d: %v", len(d), d)
	}
}
