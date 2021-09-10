package thing

import iotclient "github.com/arduino/iot-client-go"

// ThingInfo contains the main parameters of
// an Arduino IoT Cloud thing.
type ThingInfo struct {
	Name      string
	ID        string
	DeviceID  string
	Variables []string
}

func getThingInfo(thing *iotclient.ArduinoThing) *ThingInfo {
	var vars []string
	for _, p := range thing.Properties {
		vars = append(vars, p.Name)
	}
	info := &ThingInfo{
		Name:      thing.Name,
		ID:        thing.Id,
		DeviceID:  thing.DeviceId,
		Variables: vars,
	}
	return info
}