package ota

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	iotclient "github.com/arduino/iot-client-go"
)

const testFilename = "testdata/empty.bin"

type deviceUploaderTest struct {
	deviceOTA func(ctx context.Context, id string, file *os.File, expireMins int) error
}

func (d *deviceUploaderTest) DeviceOTA(ctx context.Context, id string, file *os.File, expireMins int) error {
	return d.deviceOTA(ctx, id, file, expireMins)
}

func TestRun(t *testing.T) {
	var (
		failPrefix = "00000000"
		failID1    = failPrefix + "-b39d-47a2-adf3-d26cdf474707"
		failID2    = failPrefix + "-9efd-4670-a478-df76ebdeeb4f"
		okPrefix   = "11111111"
		okID1      = okPrefix + "-4838-4f46-8930-d735c5b76cd1"
		okID2      = okPrefix + "-003f-42f9-a80c-85a1de36753b"
		okID3      = okPrefix + "-dac4-4a6a-80a4-698062fe2af5"
	)
	mockClient := &deviceUploaderTest{
		deviceOTA: func(ctx context.Context, id string, file *os.File, expireMins int) error {
			if strings.Split(id, "-")[0] == failPrefix {
				return errors.New("err")
			}
			return nil
		},
	}

	devs := []string{okID1, failID1, okID2, failID2, okID3}
	res := run(context.TODO(), mockClient, devs, testFilename, 0)
	if len(res) != len(devs) {
		t.Errorf("expected %d results, got %d", len(devs), len(res))
	}

	for _, r := range res {
		pre := strings.Split(r.ID, "-")[0]
		if pre == okPrefix && r.Err != nil {
			t.Errorf("device %s expected to succeed, but got error %s", r.ID, r.Err.Error())
		}
		if pre == failPrefix && r.Err == nil {
			t.Errorf("device %s expected to fail, but got no error", r.ID)
		}
	}
}

type deviceListerTest struct {
	list []iotclient.ArduinoDevicev2
}

func (d *deviceListerTest) DeviceList(ctx context.Context, tags map[string]string) ([]iotclient.ArduinoDevicev2, error) {
	return d.list, nil
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

	mockDeviceList := deviceListerTest{
		list: []iotclient.ArduinoDevicev2{
			{Id: idCorrect1, Fqbn: correctFQBN},
			{Id: idCorrect2, Fqbn: correctFQBN},
			{Id: idNotValid, Fqbn: wrongFQBN},
		},
	}

	ids := []string{
		idCorrect1,
		idNotFound,
		idCorrect2,
		idNotValid,
	}
	v, i, err := validateDevices(context.TODO(), &mockDeviceList, ids, correctFQBN)
	if err != nil {
		t.Errorf("unexpected error: %s", err.Error())
	}

	if len(v) != 2 {
		t.Errorf("expected 2 valid devices, but found %d: %v", len(v), v)
	}

	if len(i) != 2 {
		t.Errorf("expected 2 invalid devices, but found %d: %v", len(i), i)
	}
}
