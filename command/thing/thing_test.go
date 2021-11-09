package thing

import (
	"fmt"
	"testing"

	iotclient "github.com/arduino/iot-client-go"
)

func TestGetThingInfo(t *testing.T) {
	thingTagsValid := &iotclient.ArduinoThing{Tags: map[string]interface{}{
		"location": "rome",
		"room":     "101",
	}}
	thingTagsNotValid := &iotclient.ArduinoThing{Tags: map[string]interface{}{
		"location": "rome",
		"room":     101,
	}}

	thing, err := getThingInfo(thingTagsValid)
	if err != nil {
		t.Error("unexpected error")
	}
	if len(thing.Tags) != 2 {
		fmt.Println(len(thing.Tags))
		t.Error("expected two tags")
	}

	_, err = getThingInfo(thingTagsNotValid)
	if err == nil {
		t.Error("an error was expected because tags are not valid")
	}
}
