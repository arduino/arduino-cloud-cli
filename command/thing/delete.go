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

package thing

import (
	"context"
	"errors"

	"github.com/arduino/arduino-cloud-cli/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
)

// DeleteParams contains the parameters needed to
// delete a thing from Arduino IoT Cloud.
// ID and Tags parameters are mutually exclusive
// and one among them is required:  An error is returned
// if they are both nil or if they are both not nil.
type DeleteParams struct {
	ID   *string
	Tags map[string]string
}

// Delete command is used to delete a thing
// from Arduino IoT Cloud.
func Delete(ctx context.Context, params *DeleteParams, cred *config.Credentials) error {
	if params.ID == nil && params.Tags == nil {
		return errors.New("provide either ID or Tags")
	} else if params.ID != nil && params.Tags != nil {
		return errors.New("cannot use both ID and Tags. only one of them should be not nil")
	}

	iotClient, err := iot.NewClient(cred)
	if err != nil {
		return err
	}

	thingIDs := []string{}
	if params.ID != nil {
		thingIDs = append(thingIDs, *params.ID)
	}
	if params.Tags != nil {
		th, err := iotClient.ThingList(ctx, nil, nil, false, params.Tags)
		if err != nil {
			return err
		}
		for _, t := range th {
			thingIDs = append(thingIDs, t.Id)
		}
	}

	for _, id := range thingIDs {
		err = iotClient.ThingDelete(ctx, id)
		if err != nil {
			return err
		}
	}

	return nil
}
