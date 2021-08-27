package thing

import (
	"fmt"

	iotclient "github.com/arduino/iot-client-go"
	"github.com/arduino/iot-cloud-cli/internal/config"
	"github.com/arduino/iot-cloud-cli/internal/iot"
)

// CloneParams contains the parameters needed to clone a thing.
type CloneParams struct {
	// Mandatory - contains the name of the thing
	Name string
	// Mandatory - specifies ID of thing to be cloned
	CloneID string
}

// Clone allows to create a new thing from an already existing one
func Clone(params *CloneParams) (string, error) {
	conf, err := config.Retrieve()
	if err != nil {
		return "", err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return "", err
	}

	thing, err := retrieve(iotClient, params.CloneID)
	if err != nil {
		return "", err
	}

	thing.Name = params.Name
	force := true
	thingID, err := iotClient.AddThing(thing, force)
	if err != nil {
		return "", err
	}

	return thingID, nil
}

func retrieve(client iot.Client, thingID string) (*iotclient.Thing, error) {
	clone, err := client.GetThing(thingID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "retrieving the thing to be cloned", err)
	}

	thing := &iotclient.Thing{}

	// Copy variables
	for _, p := range clone.Properties {
		thing.Properties = append(thing.Properties, iotclient.Property{
			Name:            p.Name,
			Permission:      p.Permission,
			UpdateParameter: p.UpdateParameter,
			UpdateStrategy:  p.UpdateStrategy,
			Type:            p.Type,
			VariableName:    p.VariableName,
		})
	}

	return thing, nil
}
