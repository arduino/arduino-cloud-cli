package ping

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bcmi-labs/iot-cloud-cli/internal/mqtt"
	"github.com/bcmi-labs/iot-cloud-cli/internal/properties"
	paho "github.com/eclipse/paho.mqtt.golang"
)

type Params struct {
	Host            string
	Username        string
	Password        string
	ThingID         string
	Troubleshooting bool
}

func Ping(params *Params) error {
	if params.Troubleshooting {
		paho.ERROR = log.New(os.Stdout, "[ERROR] ", 0)
		paho.CRITICAL = log.New(os.Stdout, "[CRIT] ", 0)
		paho.WARN = log.New(os.Stdout, "[WARN]  ", 0)
		paho.DEBUG = log.New(os.Stdout, "[DEBUG] ", 0)
	}

	mqttAdapter := mqtt.NewAdapterWithAuth(
		params.Host,
		params.Username,
		params.Username,
		params.Password,
	)

	err := mqttAdapter.Connect()
	if err != nil {
		return err
	}
	fmt.Println("Connected to Arduino IoT Cloud")

	inboundTopic := fmt.Sprintf("/a/t/%s/e/i", params.ThingID)
	outboundTopic := fmt.Sprintf("/a/t/%s/e/o", params.ThingID)

	// Subscribing to the thing inbound topic to received new properties
	// values from the cloud.
	ok, _ := mqttAdapter.On(inboundTopic, func(msg paho.Message) {
		fmt.Println("received a message", msg)
	})
	fmt.Println("Subscribed", ok)

	// Sending new random values (in the 0-100 range) to the thing specified
	// using the flags
	go func() {
		for {
			randomValue := rand.Intn(100)

			property, err := properties.NewInteger("counter", randomValue)
			if err != nil {
				fmt.Println("Failed to encode property value", err)
			}

			// Publishing a new random value to the outbound topic of the thing
			err = mqttAdapter.Publish(outboundTopic, property)
			if err != nil {
				fmt.Println("Failed to send property to Arduino IoT Cloud", err)
			}
			fmt.Println("Property value sent successfully", randomValue)

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

	fmt.Println("Disconnected from Arduino IoT Cloud.")
	fmt.Println("Completed successfully.")
	return nil
}
