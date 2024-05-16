package ota

import (
	"fmt"

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/config"
	otaapi "github.com/arduino/arduino-cloud-cli/internal/ota-api"
)

func CancelOta(otaid string, cred *config.Credentials) error {

	if feedback.GetFormat() == feedback.JSONMini {
		return fmt.Errorf("jsonmini format is not supported for this command")
	}

	otapi := otaapi.NewClient(cred)

	if otaid != "" {
		_, err := otapi.CancelOta(otaid)
		if err != nil {
			return err
		}
	}

	return nil
}
