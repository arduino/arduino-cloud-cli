package thing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"errors"

	iotclient "github.com/arduino/iot-client-go"
	"github.com/arduino/iot-cloud-cli/internal/config"
	"github.com/arduino/iot-cloud-cli/internal/iot"
)

// CreateParams contains the parameters needed to create a new thing.
type CreateParams struct {
	// Mandatory - contains the name of the thing
	Name string
	// Optional - contains the ID of the device to be bound to the thing
	DeviceID string
	// Mandatory if device is empty - contains the path of the template file
	Template string
	// Mandatory if template is empty- name of things to be cloned
	CloneID string
}

// Create allows to create a new thing
func Create(params *CreateParams) (string, error) {
	if params.Template == "" && params.CloneID == "" {
		return "", fmt.Errorf("%s", "provide either a thing(ID) to clone (--clone) or a thing template file (--template)\n")
	}

	conf, err := config.Retrieve()
	if err != nil {
		return "", err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return "", err
	}

	var thing *iotclient.Thing

	if params.CloneID != "" {
		thing, err = clone(iotClient, params.CloneID)
		if err != nil {
			return "", err
		}

	} else if params.Template != "" {
		thing, err = loadTemplate(params.Template)
		if err != nil {
			return "", err
		}

	} else {
		return "", errors.New("provide either a thing(ID) to clone (--clone) or a thing template file (--template)")
	}

	thing.Name = params.Name
	force := true
	if params.DeviceID != "" {
		thing.DeviceId = params.DeviceID
	}
	thingID, err := iotClient.AddThing(thing, force)
	if err != nil {
		return "", err
	}

	return thingID, nil
}

func clone(client iot.Client, thingID string) (*iotclient.Thing, error) {
	clone, err := client.GetThing(thingID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "retrieving the thing to be cloned", err)
	}

	thing := &iotclient.Thing{}

	// Copy device id
	if clone.DeviceId != "" {
		thing.DeviceId = clone.DeviceId
	}

	// Copy properties
	for _, p := range clone.Properties {
		thing.Properties = append(thing.Properties, iotclient.Property{
			Name:            p.Name,
			MinValue:        p.MinValue,
			MaxValue:        p.MaxValue,
			Permission:      p.Permission,
			UpdateParameter: p.UpdateParameter,
			UpdateStrategy:  p.UpdateStrategy,
			Type:            p.Type,
			VariableName:    p.VariableName,
			Persist:         p.Persist,
			Tag:             p.Tag,
		})
	}

	return thing, nil
}

func loadTemplate(file string) (*iotclient.Thing, error) {
	templateFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer templateFile.Close()

	templateBytes, err := ioutil.ReadAll(templateFile)
	if err != nil {
		return nil, err
	}

	thing := &iotclient.Thing{}
	err = json.Unmarshal([]byte(templateBytes), thing)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "reading template file: template not valid: ", err)
	}

	return thing, nil
}
