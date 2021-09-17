package device

import (
	"errors"
	"fmt"

	"github.com/arduino/iot-cloud-cli/internal/config"
	"github.com/arduino/iot-cloud-cli/internal/iot"
)

// DeleteParams contains the parameters needed to
// delete a list devices from Arduino IoT Cloud.
type DeleteParams struct {
	IDs []string
}

// Delete command is used to delete devices
// from Arduino IoT Cloud.
func Delete(params *DeleteParams) error {
	conf, err := config.Retrieve()
	if err != nil {
		return err
	}
	iotClient, err := iot.NewClient(conf.Client, conf.Secret)
	if err != nil {
		return err
	}

	result := ""
	for _, id := range params.IDs {
		err = iotClient.DeviceDelete(id)
		if err != nil {
			result = fmt.Sprintf("%s\ndevice id %s: %s", result, id, err)
		}
	}

	if result != "" {
		return errors.New(result)
	}
	return nil
}
