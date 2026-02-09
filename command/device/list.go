// This file is part of arduino-cloud-cli.
//
// Copyright (C) ARDUINO SRL (http://www.arduino.cc/)
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

package device

import (
	"context"
	"fmt"
	"strings"

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
)

// ListParams contains the optional parameters needed
// to filter the devices to be listed.
type ListParams struct {
	Tags      map[string]string // If tags are provided, only devices that have all these tags are listed.
	DeviceIds string            // If ids are provided, only devices with these ids are listed.
	Status    string            // If status is provided, only devices with this status are listed.
}

// List command is used to list
// the devices of Arduino IoT Cloud.
func List(ctx context.Context, params *ListParams, cred *config.Credentials) ([]DeviceInfo, error) {
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return nil, err
	}

	foundDevices, err := iotClient.DeviceList(ctx, params.Tags)
	if err != nil {
		return nil, err
	}
	var deviceIdFilter []string
	if params.DeviceIds != "" {
		deviceIdFilter = strings.Split(params.DeviceIds, ",")
		for i := range deviceIdFilter {
			deviceIdFilter[i] = strings.TrimSpace(deviceIdFilter[i])
		}
	}

	var devices []DeviceInfo
	for _, foundDev := range foundDevices {
		if len(deviceIdFilter) > 0 && !sliceContains(deviceIdFilter, foundDev.Id) {
			continue
		}
		dev, err := getDeviceInfo(&foundDev)
		if err != nil {
			return nil, fmt.Errorf("parsing device %s from cloud: %w", foundDev.Id, err)
		}
		if params.Status != "" && dev.Status != nil && *dev.Status != params.Status {
			continue
		}
		devices = append(devices, *dev)
	}

	return devices, nil
}

func sliceContains(s []string, v string) bool {
	for i := range s {
		if v == s[i] {
			return true
		}
	}
	return false
}
