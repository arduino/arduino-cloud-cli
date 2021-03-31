package main

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/bcmi-labs/oniudra-cli/iot/codec"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/sirupsen/logrus"
	"github.com/zmoog/mariquita/utils"
)

// Create a new instance of the logger. You can have any number of instances.
var log = logrus.New()
var senmlCodec = codec.NewSenMLCodecWithLoggerString(codec.CBOR, log)

var messagePubHandler mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {

	topicName := msg.Topic()

	info, err := utils.TopicDeviceThingInfoExtractor(topicName)
	if err != nil {
		fmt.Println("cannot extract any info from", topicName)
		return
	}

	if strings.HasSuffix(topicName, "/e/i") {

		record, err := senmlCodec.Decode(msg.Payload())
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("<<< received", record)

		for _, v := range record.Values {
			fmt.Println(v)

			if strings.HasPrefix(v.Name, "counter") {

				fmt.Println("Found a counter!")

				// thing := codec.Device{
				// 	ThingID: uuid.FromStringOrNil("19e189e5-c886-4af7-94d7-0d5495c1215b"),
				// }
				// device := codec.Device{}

				if v.Type == "float" {
					// f, ok := v.Value.(float64)

					// if ok && f != 66 {
					// 	v.Value = 66

					values := codec.DevicePropertyValues{
						// Device: device,
						Values: []codec.PropertyValue{},
						// timestamp: &time.Now(),
					}

					values.AddPropertyValueNamed(v.Name, v.Value)
					fmt.Println(">>> sending", values)

					val, err := senmlCodec.Encode(values)
					if err != nil {
						fmt.Println(err)
					}

					// time.Sleep(200 * time.Millisecond)

					destinationTopicName := fmt.Sprintf("/a/t/%s/e/o", info.ID)

					fmt.Println("Val:", val, "to", destinationTopicName)
					token := client.Publish(destinationTopicName, 0, false, val)

					result := token.Wait()
					fmt.Println("publish result:", result)
				}
				// }

			}

		}
	}

	fmt.Printf("Received message: %s from topic: %s\n", msg.Payload(), msg.Topic())
}

var connectHandler mqtt.OnConnectHandler = func(client mqtt.Client) {
	fmt.Println("Connected")
}

var connectLostHandler mqtt.ConnectionLostHandler = func(client mqtt.Client, err error) {
	fmt.Printf("Connect lost: %v", err)
}

var reconnectionHandler mqtt.ReconnectHandler = func(client mqtt.Client, opts *mqtt.ClientOptions) {
	fmt.Printf("Reconnecting")
}

func main() {

	var broker = "mqtts-up.iot.arduino.cc" // prod
	// var broker = "mqtts-sa.iot.oniudra.cc" // dev
	var port = 8884

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcps://%s:%d", broker, port))

	// PUT YOUR CREDENTIALS HERE
	opts.SetClientID("")
	opts.SetUsername("")
	opts.SetPassword("")
	// PUT YOUR CREDENTIALS HERE

	opts.SetDefaultPublishHandler(messagePubHandler)
	opts.OnConnect = connectHandler
	opts.OnConnectionLost = connectLostHandler
	opts.OnReconnecting = reconnectionHandler

	tlsConfig := NewTlsConfig()
	opts.SetTLSConfig(tlsConfig)

	client := mqtt.NewClient(opts)

	fmt.Println("connecting to the broker")
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}

	for _, topic := range []string{
		// "d22c9dac-c90f-4828-83b5-522aa523ca43",
		"19e189e5-c886-4af7-94d7-0d5495c1215b", // PUT YOUR THING ID HERE
		// "1395b95e-8427-4fd2-8643-a274962d40de",
		// "19e189e5-c886-4af7-94d7-0d5495c1215b",
		// "42683b4e-58c6-43aa-853b-a523d5049f7c",
		// "44f0bb3c-4dc7-4175-8371-eafb83fcc2a7",
		// "49a4a32e-e6bb-40ce-8c81-6abbc778fa7d",
		// "ae43ba7b-fe67-4a6a-8cd5-31848655c12d",
		// "c7571f4e-77b6-418c-b892-d5313cfb2f82",
		// "d1fb8cc8-4c90-45fd-bbf5-ed72abdbefc8",
		// "d22c9dac-c90f-4828-83b5-522aa523ca43",
		// "e87b0ba2-f8ca-49b0-9c46-c61368c40dda",
		// "e8aff340-0cc6-45e7-950d-c95654da087c",
	} {
		subscribeTopicName := fmt.Sprintf("/a/t/%s/e/i", topic)
		fmt.Println("subscribing to topic: ", subscribeTopicName)
		if token := client.Subscribe(subscribeTopicName, 0, nil); token.Wait() && token.Error() != nil {
			fmt.Println(token.Error())
			panic(token.Error())
		}
	}

	fmt.Println("Ready")
	time.Sleep(20 * time.Minute)

	// bye bye
	client.Disconnect(250)
}

func NewTlsConfig() *tls.Config {
	// certpool := x509.NewCertPool()
	// ca, err := ioutil.ReadFile("ca.pem")
	// if err != nil {
	// log.Fatalln(err.Error())
	// }
	// certpool.AppendCertsFromPEM(ca)
	return &tls.Config{
		// RootCAs:    certpool,
		MinVersion: tls.VersionTLS11,
		MaxVersion: tls.VersionTLS12,
		ClientAuth: tls.NoClientCert,
		// VerifyConnection: ,
	}
}
