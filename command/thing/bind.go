package thing

import (
	"github.com/arduino/arduino-cloud-cli/internal/config"
	"github.com/arduino/arduino-cloud-cli/internal/iot"
	iotclient "github.com/arduino/iot-client-go"
)

// BindParams contains the parameters needed to
// bind a thing to a device.
type BindParams struct {
	ID       string // ID of the thing to be bound
	DeviceID string // ID of the device to be bound
}

// Bind command is used to bind a thing to a device
// on Arduino IoT Cloud.
func Bind(params *BindParams) error {
	conf, err := config.Retrieve()
	if err != nil {
		return err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return err
	}

	thing := &iotclient.Thing{
		DeviceId: params.DeviceID,
	}

	err = iotClient.ThingUpdate(params.ID, thing, true)
	if err != nil {
		return err
	}

	return nil
}
