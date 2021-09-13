package device

import (
	"github.com/arduino/iot-cloud-cli/internal/config"
	"github.com/arduino/iot-cloud-cli/internal/iot"
)

// DeviceInfo contains the most interesting
// parameters of an Arduino IoT Cloud device.
type DeviceInfo struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Board  string `json:"board"`
	Serial string `json:"serial-number"`
	FQBN   string `json:"fqbn"`
}

// List command is used to list
// the devices of Arduino IoT Cloud.
func List() ([]DeviceInfo, error) {
	conf, err := config.Retrieve()
	if err != nil {
		return nil, err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return nil, err
	}

	foundDevices, err := iotClient.DeviceList()
	if err != nil {
		return nil, err
	}

	var devices []DeviceInfo
	for _, foundDev := range foundDevices {
		dev := DeviceInfo{
			Name:   foundDev.Name,
			ID:     foundDev.Id,
			Board:  foundDev.Type,
			Serial: foundDev.Serial,
			FQBN:   foundDev.Fqbn,
		}
		devices = append(devices, dev)
	}

	return devices, nil
}
