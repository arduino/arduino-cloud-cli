package thing

import (
	"github.com/arduino/iot-cloud-cli/internal/config"
	"github.com/arduino/iot-cloud-cli/internal/iot"
)

// DeleteParams contains the parameters needed to
// delete a thing from Arduino IoT Cloud.
type DeleteParams struct {
	ID string
}

// Delete command is used to delete a thing
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

	return iotClient.DeleteThing(params.ID)
}
