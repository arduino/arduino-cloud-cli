package thing

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	iotclient "github.com/arduino/iot-client-go"
	"github.com/arduino/iot-cloud-cli/internal/config"
	"github.com/arduino/iot-cloud-cli/internal/iot"
)

// ExtractParams contains the parameters needed to
// extract a thing from Arduino IoT Cloud and save it on local storage.
// Output indicates the destination path of the extraction.
type ExtractParams struct {
	ID      string
	Outfile *string
}

// Extract command is used to extract a thing template
// from a thing on Arduino IoT Cloud.
func Extract(params *ExtractParams) error {
	conf, err := config.Retrieve()
	if err != nil {
		return err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return err
	}

	thing, err := iotClient.GetThing(params.ID)
	if err != nil {
		err = fmt.Errorf("%s: %w", "cannot extract thing: ", err)
		return err
	}

	template, err := templateFromThing(thing)
	if err != nil {
		return err
	}

	if params.Outfile == nil {
		outfile := thing.Name + "-template.json"
		params.Outfile = &outfile
	}
	err = ioutil.WriteFile(*params.Outfile, template, os.FileMode(0644))
	if err != nil {
		err = fmt.Errorf("%s: %w", "cannot write outfile: ", err)
		return err
	}

	return nil
}

func templateFromThing(thing *iotclient.ArduinoThing) ([]byte, error) {
	template := make(map[string]interface{})
	template["name"] = thing.Name

	var props []map[string]interface{}
	for _, p := range thing.Properties {
		prop := make(map[string]interface{})
		prop["name"] = p.Name
		prop["permission"] = p.Permission
		prop["type"] = p.Type
		prop["update_parameter"] = p.UpdateParameter
		prop["update_strategy"] = p.UpdateStrategy
		prop["variable_name"] = p.VariableName
		props = append(props, prop)
	}
	template["properties"] = props

	// Extract json template from thing structure
	file, err := json.MarshalIndent(template, "", "    ")
	if err != nil {
		err = fmt.Errorf("%s: %w", "thing marshal failure: ", err)
		return nil, err
	}
	return file, nil
}
