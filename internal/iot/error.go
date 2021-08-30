package iot

import (
	"fmt"

	iotclient "github.com/arduino/iot-client-go"
)

// errorDetail takes a generic iot-client-go error
// and tries to return a more detailed error.
func errorDetail(err error) error {
	apiErr, ok := err.(iotclient.GenericOpenAPIError)
	if !ok {
		return err
	}

	modErr, ok := apiErr.Model().(iotclient.ModelError)
	if !ok {
		return err
	}

	return fmt.Errorf("%w: %s", err, modErr.Detail)
}
