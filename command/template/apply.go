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
	"context"
	"fmt"
	"strings"

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	storageapi "github.com/arduino/arduino-cloud-cli/internal/storage-api"
	iotclient "github.com/arduino/iot-client-go/v2"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
)

func ApplyCustomTemplates(cred *config.Credentials, templateId, deviceId, prefix string, networkCredentials map[string]string) error {

	ctx := context.Background()

	// Open clients
	apiclient := storageapi.NewClient(cred)
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return err
	}

	feedback.Printf("Applying template %s to device %s", templateId, deviceId)

	templateIdUUID, err := uuid.FromString(templateId)
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
	mainThing := cstTemplate.ThingTemplates[0]
	logrus.Debug("Main thing template - id: ", mainThing.Id)

	// Get device and check its connectivity
	secrets, err := resolveDeviceNetworkConfigurations(ctx, iotClient, deviceId, networkCredentials)
	if err != nil {
		return err
	}
	for key, value := range secrets {
		logrus.Info(fmt.Sprintf("Secret %s: %s", key, value))
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
