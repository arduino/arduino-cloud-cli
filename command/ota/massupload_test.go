// This file is part of arduino-cloud-cli.
//
// Copyright (C) 2021 ARDUINO SA (http://www.arduino.cc/)
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
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	otaapi "github.com/arduino/arduino-cloud-cli/internal/ota-api"
	iotclient "github.com/arduino/iot-client-go/v3"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

const testFilename = "testdata/empty.bin"
const cloudFirmwareFilename = "testdata/cloud.bin"

type deviceUploaderTest struct {
	deviceOTA func(ctx context.Context, id string, file *os.File, expireMins int) error
}

func (d *deviceUploaderTest) DeviceOTA(ctx context.Context, id string, file *os.File, expireMins int) error {
	return d.deviceOTA(ctx, id, file, expireMins)
}

type otaStatusGetterTest struct{}

func (s *otaStatusGetterTest) GetOtaLastStatusByDeviceID(deviceID string) (*otaapi.OtaStatusList, error) {
	ota := otaapi.Ota{
		ID:        uuid.Must(uuid.NewV4()).String(),
		Status:    "in_progress",
		StartedAt: "2021-09-01T12:00:00Z",
	}
	response := &otaapi.OtaStatusList{
		Ota: []otaapi.Ota{ota},
	}
	return response, nil
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
	mockStatusClient := &otaStatusGetterTest{}

	devs := []string{okID1, failID1, okID2, failID2, okID3}
	res := run(context.TODO(), mockClient, mockStatusClient, devs, testFilename, 0)
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
			{Id: idCorrect1, Fqbn: &correctFQBN},
			{Id: idCorrect2, Fqbn: &correctFQBN},
			{Id: idNotValid, Fqbn: &wrongFQBN},
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

func TestValidateBuildOtaFile(t *testing.T) {

	file, tmp, err := buildOtaFile(&MassUploadParams{
		File:             cloudFirmwareFilename,
		DoNotApplyHeader: false,
		FQBN:             "arduino:samd:nano_33_iot",
	})
	assert.Nil(t, err)
	assert.NotNil(t, file)
	assert.True(t, strings.HasSuffix(file, "temp.ota"))
	assert.NotEmpty(t, tmp)
	defer os.RemoveAll(tmp)
}

func TestValidateBuildOtaFile_whenNoHeaderIsRequested(t *testing.T) {

	file, tmp, err := buildOtaFile(&MassUploadParams{
		File:             cloudFirmwareFilename,
		DoNotApplyHeader: true,
		FQBN:             "arduino:samd:nano_33_iot",
	})
	assert.Nil(t, err)
	assert.NotNil(t, file)
	assert.Equal(t, cloudFirmwareFilename, file)
	assert.Empty(t, tmp)
}
