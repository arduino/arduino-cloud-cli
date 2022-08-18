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

package tag

import (
	"context"
	"errors"

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
)

// CreateTagsParams contains the parameters needed to create or overwrite
// tags on a resource of Arduino IoT Cloud.
type CreateTagsParams struct {
	ID       string            // Resource ID
	Tags     map[string]string // Map of tags to create
	Resource ResourceType
}

// CreateTags allows to create or overwrite tags
// on a resource of Arduino IoT Cloud.
func CreateTags(ctx context.Context, params *CreateTagsParams, cred *config.Credentials) error {
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return err
	}

	switch params.Resource {
	case Thing:
		err = iotClient.ThingTagsCreate(ctx, params.ID, params.Tags)
	case Device:
		err = iotClient.DeviceTagsCreate(ctx, params.ID, params.Tags)
	default:
		err = errors.New("passed Resource parameter is not valid")
	}
	return err
}
