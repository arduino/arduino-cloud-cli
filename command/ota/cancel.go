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
		// No error, get current status
		res, err := otapi.GetOtaStatusByOtaID(otaid, 1, otaapi.OrderDesc)
		if err != nil {
			return err
		}
		if res != nil {
			feedback.PrintResult(res.Ota)
		}
		return nil
	}

	return nil
}
