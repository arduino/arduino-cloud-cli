// This file is part of arduino-cloud-cli.
//
// Copyright (C) 2024 ARDUINO SA (http://www.arduino.cc/)
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

package template

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/command/ota"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	storageapi "github.com/arduino/arduino-cloud-cli/internal/storage-api"
	iotclient "github.com/arduino/iot-client-go/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var errNoBinaryFound = errors.New("no binary found in the template")

func ApplyCustomTemplates(cred *config.Credentials, templateId, deviceId, prefix string, networkCredentials map[string]string, applyOta bool) error {

	ctx := context.Background()

	// Open clients
	apiclient := storageapi.NewClient(cred)
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return err
	}

	feedback.Printf("Applying template %s to device %s", templateId, deviceId)

	templateIdUUID, err := uuid.Parse(templateId)
	if err != nil {
		return fmt.Errorf("invalid template id: %s", templateId)
	}

	// Get custom template and verify it is present
	cstTemplate, err := apiclient.GetCustomTemplate(templateIdUUID)
	if err != nil {
		return err
	}
	if len(cstTemplate.ThingTemplates) <= 0 {
		return fmt.Errorf("template %s has no thing template", templateId)
	}
	thingTemplateIdentifier := cstTemplate.ThingTemplates[0]
	logrus.Debug("Main thing template - id: ", thingTemplateIdentifier.Id)

	// Get device and check its connectivity
	secrets, err := resolveDeviceNetworkConfigurations(ctx, iotClient, deviceId, networkCredentials)
	if err != nil {
		return err
	}

	// Apply the template
	_, err = iotClient.TemplateApply(ctx, templateId, thingTemplateIdentifier.Id, prefix, deviceId, secrets)
	if err != nil {
		return err
	}
	feedback.Printf("Template applied successfully to device %s", deviceId)

	if applyOta {
		// Now, start OTA with binary available in the template
		done, err := runOTAForTemplate(ctx, cred, templateId, deviceId, apiclient)
		if err != nil {
			return err
		}
		if done {
			feedback.Printf("OTA started successfully on device %s", deviceId)
		}
	}

	return nil
}

func resolveDeviceNetworkConfigurations(ctx context.Context, cl *iot.Client, deviceId string, networkCredentials map[string]string) (map[string]string, error) {
	device, err := cl.DeviceShow(ctx, deviceId)
	if err != nil {
		return nil, err
	}
	if device.Type == "" || device.ConnectionType == nil {
		logrus.Warnf("Device %s has no type or connection-type - type: %s", deviceId, device.Type)
		return nil, nil // cannot take a decision on this device, try to proceed
	}
	logrus.Infof("Device %s - type: %s - connection-type: %s", deviceId, device.Type, *device.ConnectionType)

	// Check if device is linked to a thing. In such case, block the operation.
	if device.Thing != nil && device.Thing.Id != "" {
		return nil, fmt.Errorf("device %s is already linked to a thing (thing_id: %s)", deviceId, device.Thing.Id)
	}

	credentials, err := cl.DeviceNetworkCredentials(ctx, device.Type, *device.ConnectionType)
	if err != nil {
		return nil, err
	}

	// Check if the provided network credentials are valid. Verify if all the required credentials are present.
	discoveredCredentials := make(map[string]iotclient.ArduinoCredentialsv1)
	for _, credential := range credentials {
		discoveredCredentials[credential.GetSecretName()] = credential
		if credential.Required {
			if _, ok := networkCredentials[credential.GetSecretName()]; !ok {
				return nil, fmt.Errorf("missing mandatory network credential: %s. Available: %s", credential.GetSecretName(), humanReadableCredentials(credentials))
			}
		}
	}
	// Remove any property that is not supported
	for key := range networkCredentials {
		if _, ok := discoveredCredentials[key]; !ok {
			delete(networkCredentials, key)
		}
	}

	return networkCredentials, nil
}

func humanReadableCredentials(cred []iotclient.ArduinoCredentialsv1) string {
	var buf strings.Builder
	for _, c := range cred {
		if c.Required {
			buf.WriteString(fmt.Sprintf("  - %s (required)", c.GetSecretName()))
		} else {
			buf.WriteString(fmt.Sprintf("  - %s", c.GetSecretName()))
		}
	}
	return buf.String()
}

func runOTAForTemplate(ctx context.Context, cred *config.Credentials, templateId, deviceId string, apiclient *storageapi.StorageApiClient) (bool, error) {
	otaTempDir, err := os.MkdirTemp("cli-template-ota", "")
	if err != nil {
		return false, fmt.Errorf("%s: %w", "cannot create temporary folder", err)
	}
	defer func() {
		err := os.RemoveAll(otaTempDir)
		if err != nil {
			logrus.Warnf("Failed to remove temp directory: %v", err)
		}
	}()

	filecreaed, err := apiclient.ExportCustomTemplate(templateId, otaTempDir)
	if err != nil {
		return false, err
	}

	// open the file and be ready to send it to the device
	otaFile, err := extractBinary(filecreaed, otaTempDir)
	if err != nil {
		return false, err
	}
	if otaFile == "" {
		feedback.Printf("No binary OTA file found in the template")
		return false, nil
	}

	// Upload the OTA file to the device
	err = ota.Upload(ctx, &ota.UploadParams{
		DeviceID: deviceId,
		File:     otaFile,
	}, cred)
	if err != nil {
		return false, err
	}

	return true, nil
}

func extractBinary(filepath *string, tempDir string) (string, error) {
	zipFile, err := os.Open(*filepath)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	fileInfo, err := zipFile.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	zipReader, err := zip.NewReader(zipFile, fileInfo.Size())
	if err != nil {
		return "", fmt.Errorf("failed to open archive reader: %w", err)
	}

	var binaryFile *zip.File
	for _, file := range zipReader.File {
		if strings.Contains(file.Name, "resources/binaries") {
			logrus.Debugf("binary OTA file from template: %s", file.Name)
			binaryFile = file
			break
		}
	}

	if binaryFile != nil {
		// Extract content to a temporary file
		tempFile, err := os.CreateTemp(tempDir, "tmpl_bin_*.bin")
		if err != nil {
			return "", fmt.Errorf("failed to create temporary file: %w", err)
		}
		defer tempFile.Close()
		inputF, err := binaryFile.Open()
		if err != nil {
			return "", fmt.Errorf("failed to open file in archive: %w", err)
		}
		defer inputF.Close()

		_, err = io.Copy(tempFile, inputF)
		if err != nil {
			return "", fmt.Errorf("failed to copy file content: %w", err)
		}

		return tempFile.Name(), nil
	}

	return "", errNoBinaryFound
}
