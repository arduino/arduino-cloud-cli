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

package device

import (
	"context"
	"fmt"

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
)

// FrequencyPlanInfo describes a LoRa frequency plan.
type FrequencyPlanInfo struct {
	Name     string `json:"name"`
	ID       string `json:"id"`
	Advanced string `json:"advanced"`
}

// ListFrequencyPlans command is used to list
// the supported LoRa frequency plans.
func ListFrequencyPlans(ctx context.Context, cred *config.Credentials) ([]FrequencyPlanInfo, error) {
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return nil, err
	}

	foundFreqs, err := iotClient.LoraFrequencyPlansList(ctx)
	if err != nil {
		return nil, err
	}

	freqs := make([]FrequencyPlanInfo, 0, len(foundFreqs))
	for _, f := range foundFreqs {
		freq := FrequencyPlanInfo{
			Name:     f.Name,
			ID:       f.Id,
			Advanced: fmt.Sprintf("%v", f.Advanced),
		}
		freqs = append(freqs, freq)
	}

	return freqs, nil
}
