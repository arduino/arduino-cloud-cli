package thing

import (
	"github.com/arduino/iot-cloud-cli/internal/config"
	"github.com/arduino/iot-cloud-cli/internal/iot"
)

// ThingInfo contains the main parameters of
// an Arduino IoT Cloud thing.
type ThingInfo struct {
	Name       string
	ID         string
	DeviceID   string
	Properties []string
}

// ListParams contains the optional parameters needed
// to filter the things to be listed.
// If IDs is valid, only things belonging to that list are listed.
// If DeviceID is provided, only things associated to that device are listed.
// If Properties is true, properties names are retrieved.
type ListParams struct {
	IDs        []string
	DeviceID   *string
	Properties bool
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

	foundThings, err := iotClient.ListThings(params.IDs, params.DeviceID, params.Properties)
	if err != nil {
		return nil, err
	}

	var things []ThingInfo
	for _, foundThing := range foundThings {
		var props []string
		for _, p := range foundThing.Properties {
			props = append(props, p.Name)
		}
		th := ThingInfo{
			Name:       foundThing.Name,
			ID:         foundThing.Id,
			DeviceID:   foundThing.DeviceId,
			Properties: props,
		}
		things = append(things, th)
	}

	return things, nil
}
