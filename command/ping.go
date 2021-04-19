package command

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bcmi-labs/iot-cloud-cli/adapters/mqtt"
	"github.com/bcmi-labs/iot-cloud-cli/internal/properties"
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
)

var (
	host            string
	username        string
	password        string
	thingID         string
	troubleshooting bool
)

func init() {
	pingCommand.Flags().StringVarP(&host, "host", "b", "tcps://mqtts-up.iot.arduino.cc:8884", "Broker endpoint (required)")
	pingCommand.Flags().StringVarP(&username, "username", "u", "", "Username (required)")
	pingCommand.Flags().StringVarP(&password, "password", "p", "", "Password (required)")
	pingCommand.Flags().StringVarP(&thingID, "thing_id", "t", "", "Thing ID (required)")

	pingCommand.Flags().BoolVar(&troubleshooting, "troubleshooting", false, "Enable troubleshooting mode (full logs from the MQTT client)")

	pingCommand.MarkFlagRequired("username")
	pingCommand.MarkFlagRequired("password")
	pingCommand.MarkFlagRequired("thing_id")

	RootCmd.AddCommand(pingCommand)
}

var pingCommand = &cobra.Command{
	Use:   "ping",
	Short: "Ping Arduino IoT Cloud",
	Long:  "Ping Arduino IoT Cloud",
	RunE: func(cmd *cobra.Command, args []string) error {

		if troubleshooting {
			paho.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
			paho.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
			paho.WARN = log.New(os.Stdout, "[WARN]  ", 0)
			paho.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)
		}

		mqttAdapter := mqtt.NewAdapterWithAuth(
			host,
			username,
			username,
			password,
		)

		err := mqttAdapter.Connect()
		if err != nil {
			return err
		}
		fmt.Println(" * Connected to Arduino IoT Cloud")

		inboundTopic := fmt.Sprintf("/a/t/%s/e/i", thingID)
		outboundTopic := fmt.Sprintf("/a/t/%s/e/o", thingID)

		// Subscribing to the thing inbound topic to received new properties
		// values from the cloud.
		ok, _ := mqttAdapter.On(inboundTopic, func(msg paho.Message) {
			propertyValue, err := properties.From(msg.Payload())
			if err != nil {
				fmt.Println(" ! failed to decode nessage from", inboundTopic)
			}

			fmt.Println(" < received property value", propertyValue.Values)
		})
		if ok {
			fmt.Println(" * Subscribed to topic", inboundTopic)
		} else {
			fmt.Println(" ! Failed to subscribe to topic", inboundTopic)
		}

		// Sending new random values (in the 0-100 range) to the thing specified
		// using the flags
		go func() {
			for {
				randomValue := rand.Intn(100)

				property, err := properties.NewInteger("counter", randomValue)
				if err != nil {
					fmt.Println(" ! Failed to encode property value", err)
				}

				// Publishing a new random value to the outbound topic of the thing
				err = mqttAdapter.Publish(outboundTopic, property)
				if err != nil {
					fmt.Println(" ! Failed to send property to Arduino IoT Cloud", err)
				}
				fmt.Println(" > sent property value", randomValue)

				wait := 3
				time.Sleep(time.Duration(wait) * time.Second)
			}
		}()

		// Wait for a sigterm signal (CTRL+C) to disconnect from the broker
		// and say good bye
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

		// blocking here waiting for a signal from the terminal 😪
		<-c

		mqttAdapter.Disconnect()
		if err != nil {
			return err
		}
		fmt.Println(" * Disconnected from Arduino IoT Cloud.")
		fmt.Println(" * Completed successfully.")

		return nil
	},
}
