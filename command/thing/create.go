package thing

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	iotclient "github.com/arduino/iot-client-go"
	"github.com/arduino/iot-cloud-cli/internal/config"
	"github.com/arduino/iot-cloud-cli/internal/iot"
)

// CreateParams contains the parameters needed to create a new thing.
type CreateParams struct {
	// Optional - contains the name of the thing
	Name string
	// Mandatory - contains the path of the template file
	Template string
}

// Create allows to create a new thing
func Create(params *CreateParams) (string, error) {
	conf, err := config.Retrieve()
	if err != nil {
		return "", err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return "", err
	}

	thing, err := loadTemplate(params.Template)
	if err != nil {
		return "", err
	}

	// Name passed as parameter has priority over name from template
	if params.Name != "" {
		thing.Name = params.Name
	}
	// If name is not specified in the template, it should be passed as parameter
	if thing.Name == "" {
		return "", errors.New("thing name not specified")
	}

	force := true
	thingID, err := iotClient.AddThing(thing, force)
	if err != nil {
		return "", err
	}

	return thingID, nil
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

	template := make(map[string]interface{})
	err = json.Unmarshal([]byte(templateBytes), &template)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "reading template file: template not valid: ", err)
	}

	// Adapt thing template to thing structure
	delete(template, "id")
	template["properties"] = template["variables"]
	delete(template, "variables")

	// Convert template into thing structure exploiting json marshalling/unmarshalling
	thing := &iotclient.Thing{}

	t, err := json.Marshal(template)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "extracting template", err)
	}

	err = json.Unmarshal(t, &thing)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", "creating thing structure from template", err)
	}

	return thing, nil
}
