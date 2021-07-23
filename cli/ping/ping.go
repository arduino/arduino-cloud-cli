package ping

import (
	"github.com/bcmi-labs/iot-cloud-cli/command/ping"
	"github.com/spf13/cobra"
)

var (
	host            string
	device          string
	secret          string
	thing           string
	troubleshooting bool
)

func NewCommand() *cobra.Command {
	pingCommand := &cobra.Command{
		Use:   "ping",
		Short: "Ping Arduino IoT Cloud",
		Long:  "Ping Arduino IoT Cloud",
		RunE:  runPingCommand,
	}

	pingCommand.Flags().StringVarP(&host, "host", "b", "tcps://mqtts-up.iot.arduino.cc:8884", "MQTT Broker URL")
	pingCommand.Flags().StringVarP(&device, "device", "d", "", "Device ID")
	pingCommand.Flags().StringVarP(&secret, "secret", "s", "", "Device Secret")
	pingCommand.Flags().StringVarP(&thing, "thing", "t", "", "Thing ID")

	pingCommand.Flags().BoolVar(&troubleshooting, "troubleshooting", false, "Enable troubleshooting mode (full logs from the MQTT client)")

	pingCommand.MarkFlagRequired("device")
	pingCommand.MarkFlagRequired("secret")
	pingCommand.MarkFlagRequired("thing")

	return pingCommand
}

func runPingCommand(cmd *cobra.Command, args []string) error {
	params := &ping.Params{
		Host:            host,
		Username:        device,
		Password:        secret,
		ThingID:         thing,
		Troubleshooting: troubleshooting,
	}

	return ping.Ping(params)
}
