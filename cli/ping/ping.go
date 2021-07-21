package ping

import (
	"github.com/bcmi-labs/iot-cloud-cli/command/ping"
	"github.com/spf13/cobra"
)

var (
	host            string
	username        string
	password        string
	thingID         string
	troubleshooting bool
)

func NewCommand() *cobra.Command {
	pingCommand := &cobra.Command{
		Use:   "ping",
		Short: "Ping Arduino IoT Cloud",
		Long:  "Ping Arduino IoT Cloud",
		RunE:  runPingCommand,
	}

	pingCommand.Flags().StringVarP(&host, "host", "b", "tcps://mqtts-up.iot.arduino.cc:8884", "Broker endpoint (required)")
	pingCommand.Flags().StringVarP(&username, "username", "u", "", "Username (required)")
	pingCommand.Flags().StringVarP(&password, "password", "p", "", "Password (required)")
	pingCommand.Flags().StringVarP(&thingID, "thing_id", "t", "", "Thing ID (required)")

	pingCommand.Flags().BoolVar(&troubleshooting, "troubleshooting", false, "Enable troubleshooting mode (full logs from the MQTT client)")

	pingCommand.MarkFlagRequired("username")
	pingCommand.MarkFlagRequired("password")
	pingCommand.MarkFlagRequired("thing_id")

	return pingCommand
}

func runPingCommand(cmd *cobra.Command, args []string) error {
	params := &ping.PingParams{
		Host:            host,
		Username:        username,
		Password:        password,
		ThingID:         thingID,
		Troubleshooting: troubleshooting,
	}

	return ping.Ping(params)
}
