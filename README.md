# iot-cloud-cli

iot-cloud-cli is a command line interface that allows to exploit the features of Arduino IoT Cloud. As of now, it is possible to provision a device and to simulate a device to be connected to the cloud.

### Requirements

This is all you need to use iot-cloud-cli for device **provisioning**:
 * A client ID and a secret ID, retrievable from the [cloud](https://create.arduino.cc/iot/integrations) by creating a new API key
 * arduino-cli in daemon mode, use the command 'arduino-cli daemon'

This is all you need to use iot-cloud-cli as a **virtual device**:
 * A "Generic ESP8266 Module" device in IoT Cloud (requires a Maker plan)
 * A thing with a `counter` property connected to the "Generic ESP8266 Module" device 


## Set a configuration

iot-cloud-cli should be configured before being used. In particular a client ID and the corresponding secret ID should be set.
You can retrieve them from the [cloud](https://create.arduino.cc/iot/integrations) by creating a new API key.

Once you have the IDs, call this command with your parameters:

`$ iot-cloud-cli config -c <clientID> -s <secretID>`

## Device provisioning

When provisioning a device, you can optionally explicit the port which the device is connected to and its fqbn. If they are not given, then the first device found is provisioned.
Use this command to provision a device:

`$ iot-cloud-cli device create --name <deviceName> --port <port> --fqbn <deviceFqbn>`


## Use iot-cloud-cli as a virtual device

The iot-cloud-cli can be used as a virtual device for Arduino IoT Cloud for testing.

```
$ iot-cloud-cli ping -u "<deviceId>" -p "<secret>" -t <thing ID>>
  Connected to Arduino IoT Cloud
  Subscribed true
  Property value sent successfully 81
  Property value sent successfully 87
```

### How to set up the device and thing in IoT Cloud

#### Device

 * Visit https://create.arduino.cc/iot/devices and select "Add device".
 * Select "Set up a 3rd party device".
 * Select "ESP8266". 
 * From the drop down select "Generic ESP8266 Module", and click "Continue".
 * Pick a nice and friendly device name.
 * Save the "Device ID" and "Secret Key" is a safe place, because you will not be able to see them anymore.
  
#### Thing ID

 * Visit https://create.arduino.cc/iot/things and select "Create Thing".
 * Select "Add Variable".
 * Give the variable the name "counter", type "Integer Number" and leave the variable permission the value "Read & Write".
 * Press the "Add Variable" button to confirm.
 * Copy the "Thing ID" from the bottom right of the page.
 
#### Connect the device and the thing

You should connect the new device to the new thing.

#### Testing

##### Connect to the PROD environment

```shell
$ iot-cloud-cli ping -u "<Device ID>" -p "<Secret Key>" -t <Thing ID>>
```

If every works as expected you should see something similar to this output:
```
Connected to Arduino IoT Cloud
Subscribed true
Property value sent successfully 81
Property value sent successfully 87
```

If you visit https://create.arduino.cc/iot/devices the "Generic ESP8266 Module" device status should be "Online".

##### Connect to the DEV environment

The DEV environment is using a different broker, so you need to add the option `--host`:

```shell
$ iot-cloud-cli ping --host tcps://mqtts-sa.iot.oniudra.cc:8884 -u "<Device ID>" -p "<Secret Key>" -t "<thing-id>"
```
