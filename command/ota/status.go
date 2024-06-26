package ota

import (
	"fmt"

	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cloud-cli/config"
	otaapi "github.com/arduino/arduino-cloud-cli/internal/ota-api"
)

func PrintOtaStatus(otaid, otaids, device string, cred *config.Credentials, limit int, order string) error {

	if feedback.GetFormat() == feedback.JSONMini {
		return fmt.Errorf("jsonmini format is not supported for this command")
	}

	otapi := otaapi.NewClient(cred)

	if otaids != "" {
		res, err := otapi.GetOtaStatusByOtaIDs(otaids)
		if err == nil && res != nil {
			feedback.PrintResult(res)
		} else if err != nil {
			return err
		}
	} else if otaid != "" {
		res, err := otapi.GetOtaStatusByOtaID(otaid, limit, order)
		if err == nil && res != nil {
			feedback.PrintResult(otaapi.OtaStatusDetail{
				FirmwareSize: res.FirmwareSize,
				Ota:          res.Ota,
				Details:      res.States,
			})
		} else if err != nil {
			return err
		}
	} else if device != "" {
		res, err := otapi.GetOtaStatusByDeviceID(device, limit, order)
		if err == nil && res != nil {
			feedback.PrintResult(res)
		} else if err != nil {
			return err
		}
	}

	return nil
}
