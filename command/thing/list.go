package thing

import (
	"github.com/arduino/arduino-cloud-cli/internal/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
)

// ListParams contains the optional parameters needed
// to filter the things to be listed.
type ListParams struct {
	IDs       []string // If IDs is not nil, only things belonging to that list are returned
	DeviceID  *string  // If DeviceID is provided, only the thing associated to that device is listed.
	Variables bool     // If Variables is true, variable names are retrieved.
}

// List command is used to list
// the things of Arduino IoT Cloud.
func List(params *ListParams) ([]ThingInfo, error) {
	conf, err := config.Retrieve()
	if err != nil {
		return nil, err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return nil, err
	}

	foundThings, err := iotClient.ThingList(params.IDs, params.DeviceID, params.Variables)
	if err != nil {
		return nil, err
	}

	var things []ThingInfo
	for _, foundThing := range foundThings {
		info := getThingInfo(&foundThing)
		things = append(things, *info)
	}

	return things, nil
}
