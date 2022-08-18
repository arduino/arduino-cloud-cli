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

// DeleteTagsParams contains the parameters needed to
// delete tags of a device from Arduino IoT Cloud.
type DeleteTagsParams struct {
	ID       string
	Keys     []string // Keys of tags to delete
	Resource ResourceType
}

// DeleteTags command is used to delete tags of a device
// from Arduino IoT Cloud.
func DeleteTags(ctx context.Context, params *DeleteTagsParams, cred *config.Credentials) error {
	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return err
	}

	switch params.Resource {
	case Thing:
		err = iotClient.ThingTagsDelete(ctx, params.ID, params.Keys)
	case Device:
		err = iotClient.DeviceTagsDelete(ctx, params.ID, params.Keys)
	default:
		err = errors.New("passed Resource parameter is not valid")
	}
	return err
}
