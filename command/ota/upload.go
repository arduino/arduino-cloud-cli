package ota

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/arduino/iot-cloud-cli/internal/config"
	"github.com/arduino/iot-cloud-cli/internal/iot"
)

const (
	// default ota should complete in 10 mins
	otaExpirationMins = 10
	// deferred ota can take up to 1 week (equal to 10080 minutes)
	otaDeferredExpirationMins = 10080
)

// UploadParams contains the parameters needed to
// perform an OTA upload.
type UploadParams struct {
	DeviceID string
	File     string
	Deferred bool
}

// Upload command is used to upload a firmware OTA,
// on a device of Arduino IoT Cloud.
func Upload(params *UploadParams) error {
	conf, err := config.Retrieve()
	if err != nil {
		return err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return err
	}

	dev, err := iotClient.DeviceShow(params.DeviceID)
	if err != nil {
		return err
	}

	otaDir, err := ioutil.TempDir("", "")
	if err != nil {
		return fmt.Errorf("%s: %w", "cannot create temporary folder", err)
	}
	otaFile := filepath.Join(otaDir, "temp.ota")
	defer os.RemoveAll(otaDir)

	err = Generate(params.File, otaFile, dev.Fqbn)
	if err != nil {
		return fmt.Errorf("%s: %w", "cannot generate .ota file", err)
	}

	file, err := os.Open(otaFile)
	if err != nil {
		return fmt.Errorf("%s: %w", "cannot open ota file", err)
	}

	expiration := otaExpirationMins
	if params.Deferred {
		expiration = otaDeferredExpirationMins
	}

	err = iotClient.DeviceOTA(params.DeviceID, file, expiration)
	if err != nil {
		return err
	}

	return nil
}
