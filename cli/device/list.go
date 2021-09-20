package device

import (
	"os"

	"github.com/arduino/arduino-cli/cli/errorcodes"
	"github.com/arduino/arduino-cli/cli/feedback"
	"github.com/arduino/arduino-cli/table"
	"github.com/arduino/arduino-cloud-cli/command/device"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func initListCommand() *cobra.Command {
	listCommand := &cobra.Command{
		Use:   "list",
		Short: "List devices",
		Long:  "List devices on Arduino IoT Cloud",
		Run:   runListCommand,
	}
	return listCommand
}

func runListCommand(cmd *cobra.Command, args []string) {
	logrus.Info("Listing devices")

	devs, err := device.List()
	if err != nil {
		feedback.Errorf("Error during device list: %v", err)
		os.Exit(errorcodes.ErrGeneric)
	}

	feedback.PrintResult(listResult{devs})
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
