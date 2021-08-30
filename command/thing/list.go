package thing

import (
	"github.com/arduino/iot-cloud-cli/internal/config"
	"github.com/arduino/iot-cloud-cli/internal/iot"
)

// ThingInfo contains the main parameters of
// an Arduino IoT Cloud thing.
type ThingInfo struct {
	Name      string
	ID        string
	DeviceID  string
	Variables []string
}

// ListParams contains the optional parameters needed
// to filter the things to be listed.
// If IDs is valid, only things belonging to that list are listed.
// If DeviceID is provided, only things associated to that device are listed.
// If Variables is true, variables names are retrieved.
type ListParams struct {
	IDs       []string
	DeviceID  *string
	Variables bool
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
		var vars []string
		for _, p := range foundThing.Properties {
			vars = append(vars, p.Name)
		}
		th := ThingInfo{
			Name:      foundThing.Name,
			ID:        foundThing.Id,
			DeviceID:  foundThing.DeviceId,
			Variables: vars,
		}
		things = append(things, th)
	}

	return things, nil
}
