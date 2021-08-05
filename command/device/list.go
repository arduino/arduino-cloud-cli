package device

import (
	"github.com/arduino/iot-cloud-cli/internal/config"
	"github.com/arduino/iot-cloud-cli/internal/iot"
)

type DeviceInfo struct {
	Name   string
	ID     string
	Board  string
	Serial string
	FQBN   string
}

func List() ([]DeviceInfo, error) {
	conf, err := config.Retrieve()
	if err != nil {
		return nil, err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return nil, err
	}

	foundDevices, err := iotClient.ListDevices()
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
