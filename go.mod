module github.com/bcmi-labs/iot-cloud-cli

go 1.15

require (
	github.com/arduino/iot-client-go v1.3.3
	github.com/bcmi-labs/oniudra-cli v0.15.8
	github.com/eclipse/paho.mqtt.golang v1.3.2
	github.com/juju/errors v0.0.0-20200330140219-3fe23663418f
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
)

replace github.com/cisco/senml => github.com/bcmi-labs/senml v0.1.0
