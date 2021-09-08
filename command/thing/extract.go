package thing

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	iotclient "github.com/arduino/iot-client-go"
	"github.com/arduino/iot-cloud-cli/internal/config"
	"github.com/arduino/iot-cloud-cli/internal/iot"
	"gopkg.in/yaml.v3"
)

// ExtractParams contains the parameters needed to
// extract a thing from Arduino IoT Cloud and save it on local storage.
// Format determines the file format of the template ("json" or "yaml")
// Output indicates the destination path of the extraction.
type ExtractParams struct {
	ID      string
	Format  string
	Outfile *string
}

// Extract command is used to extract a thing template
// from a thing on Arduino IoT Cloud.
func Extract(params *ExtractParams) error {
	params.Format = strings.ToLower(params.Format)
	if params.Format != "json" && params.Format != "yaml" {
		return errors.New("format is not valid: only 'json' and 'yaml' are supported")
	}

	conf, err := config.Retrieve()
	if err != nil {
		return err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return err
	}

	thing, err := iotClient.ThingShow(params.ID)
	if err != nil {
		err = fmt.Errorf("%s: %w", "cannot extract thing: ", err)
		return err
	}

	template, err := templateFromThing(thing)
	if err != nil {
		return err
	}

	err = templateToFile(template, params)
	if err != nil {
		return err
	}

	return nil
}

func templateFromThing(thing *iotclient.ArduinoThing) (map[string]interface{}, error) {
	template := make(map[string]interface{})
	template["name"] = thing.Name

	// Extract template from thing structure
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
	template["variables"] = props

	return template, nil
}

func templateToFile(template map[string]interface{}, params *ExtractParams) error {
	var file []byte
	var err error

	if params.Format == "json" {
		file, err = json.MarshalIndent(template, "", "    ")
		if err != nil {
			return fmt.Errorf("%s: %w", "thing marshal failure: ", err)
		}

	} else if params.Format == "yaml" {
		file, err = yaml.Marshal(template)
		if err != nil {
			return fmt.Errorf("%s: %w", "thing marshal failure: ", err)
		}

	} else {
		return errors.New("format is not valid: only 'json' and 'yaml' are supported")
	}

	if params.Outfile == nil {
		name, ok := template["name"].(string)
		if name == "" || !ok {
			return errors.New("thing template does not have a valid name")
		}
		outfile := name + "." + params.Format
		params.Outfile = &outfile
	}

	err = ioutil.WriteFile(*params.Outfile, file, os.FileMode(0644))
	if err != nil {
		return fmt.Errorf("%s: %w", "cannot write outfile: ", err)
	}

	return nil
}