module github.com/arduino/iot-cloud-cli

go 1.15

require (
	github.com/arduino/arduino-cli v0.0.0-20210607095659-16f41352eac3
	github.com/arduino/go-paths-helper v1.6.0
	github.com/arduino/iot-client-go v1.3.3
	github.com/bcmi-labs/oniudra-cli v0.15.8
	github.com/eclipse/paho.mqtt.golang v1.3.2
	github.com/howeyc/crc16 v0.0.0-20171223171357-2b2a61e366a6
	github.com/juju/errors v0.0.0-20200330140219-3fe23663418f
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.6.1
	go.bug.st/serial v1.3.0
	golang.org/x/net v0.0.0-20210505024714-0287a6fb4125 // indirect
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	golang.org/x/sys v0.0.0-20210503173754-0981d6026fa6 // indirect
	google.golang.org/genproto v0.0.0-20210504143626-3b2ad6ccc450 // indirect
	google.golang.org/grpc v1.39.0
)

replace github.com/cisco/senml => github.com/bcmi-labs/senml v0.1.0
