package device

import (
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/iot-cloud-cli/command/device"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List devices",
		Long:  "List devices on Arduino IoT Cloud",
		RunE:  runListCommand,
	}
	return listCommand
}

func runListCommand(cmd *cobra.Command, args []string) error {
	logrus.Info("Listing devices")

	devs, err := device.List()
	if err != nil {
		return err
	}

	feedback.PrintResult(listResult{devs})

	return nil
}

type listResult struct {
	devices []device.DeviceInfo
}

func (r listResult) Data() interface{} {
	return r.devices
}

func (r listResult) String() string {
	if len(r.devices) == 0 {
		return "No devices found."
	}
	t := table.New()
	t.SetHeader("Name", "ID", "Board", "FQBN", "SerialNumber")
	for _, device := range r.devices {
		t.AddRow(
			device.Name,
			device.ID,
			device.Board,
			device.FQBN,
			device.Serial,
		)
	}
	return t.Render()
}
