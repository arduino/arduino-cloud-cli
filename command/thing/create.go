package thing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/arduino/iot-cloud-cli/internal/config"
	"github.com/arduino/iot-cloud-cli/internal/iot"
)

// CreateParams contains the parameters needed to create a new thing.
type CreateParams struct {
	// Mandatory - contains the name of the thing
	Name string
	// Optional - contains the ID of the device to be bound to the thing
	Device string
	// Mandatory if device is empty - contains the path of the template file
	Template string
	// Mandatory if template is empty- name of things to be cloned
	Clone string
}

// Create allows to create a new thing
func Create(params *CreateParams) (string, error) {
	if params.Template == "" && params.Clone == "" {
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

	var thing map[string]interface{}

	if params.Clone != "" {
		thing, err = cloneThing(iotClient, params.Clone)
		if err != nil {
			return "", err
		}

	} else if params.Template != "" {
		thing, err = loadTemplate(params.Template)
		if err != nil {
			return "", err
		}
	}

	thing["name"] = params.Name
	force := true
	if params.Device != "" {
		thing["device_id"] = params.Device
	}
	thingID, err := iotClient.AddThing(thing, force)
	if err != nil {
		return "", err
	}

	return thingID, nil
}

func cloneThing(client iot.Client, thingID string) (map[string]interface{}, error) {
	clone, err := client.GetThing(thingID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "retrieving the thing to be cloned", err)
	}

	thing := make(map[string]interface{})
	if clone.DeviceId != "" {
		thing["device_id"] = clone.DeviceId
	}
	if clone.Properties != nil {
		thing["properties"] = clone.Properties
	}

	return thing, nil
}

func loadTemplate(file string) (map[string]interface{}, error) {
	templateFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer templateFile.Close()

	templateBytes, err := ioutil.ReadAll(templateFile)
	if err != nil {
		return nil, err
	}

	var template map[string]interface{}
	err = json.Unmarshal([]byte(templateBytes), &template)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "reading template file: template not valid: ", err)
	}

	return template, nil
}
