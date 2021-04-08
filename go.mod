module github.com/bcmi-labs/iot-cloud-cli

go 1.15

require (
	github.com/bcmi-labs/oniudra-cli v0.15.8
	github.com/eclipse/paho.mqtt.golang v1.3.2
	github.com/juju/errors v0.0.0-20200330140219-3fe23663418f
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
)

replace github.com/cisco/senml => github.com/bcmi-labs/senml v0.1.0
