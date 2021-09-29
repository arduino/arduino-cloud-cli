# arduino-cloud-cli

arduino-cloud-cli is a command line tool that empowers users to access the major features of Arduino IoT Cloud from a terminal. 

### License
This code is licensed under the terms of the GNU Affero General Public License v3.0. If you have questions about licensing or need a commercial license please contact us at [license@arduino.cc](mailto:license@arduino.cc).

### Requirements

This is all you need to use arduino-cloud-cli for device **provisioning**:
 * A client ID and a secret ID, retrievable from the [cloud](https://create.arduino.cc/iot/integrations) by creating a new API key
 * The folder containing the precompiled provisioning firmwares (`binaries`) needs to be in the same location you run the command from

### Additional info

This tool follows a "quiet on success/verbose on error" behaviour. This means that when the execution of a command results in an error, such error is printed. On the other hand, when the command is successfully executed, there is no 'success' feedback on the output: the results of the command, if any, are directly printed without any other feedback informations. This strategy allows users to save the output of the command in a file. 

However, if the verbose flag `-v` is used, then the behaviour will change: the logs will always be printed out providing users with feedbacks on the execution of the command. 

As an example, let's take the `device create`command. We want to save the information of the newly created device in a file.
So we simply lunch the command:

`$ arduino-cloud-cli device create --name mydevice --format json > mydevice.json`

The resulting mydevice.json will only contain device information in a json format.

Another example: let's say that the execution of the previous command results in an error. In that case the json file will be empty and the terminal will print out the error. Now we want to execute again the command with the logs enabled, in order to understand the issue. So we execute the following:

`$ arduino-cloud-cli device create --name mydevice -v`


## Set a configuration

arduino-cloud-cli needs a configuration file to be used. At the moment, the configuration file should be contained in the same directory where the cli commands are executed.
The configuration file contains the Arduino IoT Cloud client ID and its corresponding secret.
You can retrieve them from the [cloud](https://create.arduino.cc/iot/integrations) by creating a new API key.

Once you have the IDs, call this command to init a new configuration file:

`$ arduino-cloud-cli config init`

A file named `arduino-cloud.yaml` will be created in the current working directory. 
Then you should open such file and replace the client and secret placeholders with the value you previously retrieved.


To create a configuration file in a different folder, use this command:

`$ arduino-cloud-cli config init --dest-dir <destinationFolder>`

To reset an old configuration file, just overwrite it using this command:

`$ arduino-cloud-cli config init --overwrite`

Configuration file is supported in two different format: json and yaml. Use the `--config-format` to choose it. Default is yaml.

`$ arduino-cloud-cli config init --config-format json`

## Device provisioning

When provisioning a device, you can optionally specify the port to which the device is connected to and its fqbn. If they are not given, then the first device found will be provisioned.

Use this command to provision a device:

`$ arduino-cloud-cli device create --name <deviceName> --port <port> --fqbn <deviceFqbn>`

## Device commands

Once a device has been created thorugh the provisioning procedure, it can be deleted by using the following command:
`$ arduino-cloud-cli device delete --id <deviceID>`

Devices currently present on Arduino IoT Cloud can be retrieved by using this command:
`$ arduino-cloud-cli device list`

## Thing commands

Things can be created starting from a template or by cloning another thing.

Create a thing from a thing template. Supported template formats are JSON and YAML. The name parameter is optional. If it is provided then it overrides the name retrieved from the template:

`$ arduino-cloud-cli thing create --name <thingName> --template <template.(json|yaml)>`

Create a thing by cloning another thing, here the *name is mandatory*:

`$ arduino-cloud-cli thing clone --name <thingName> --clone-id <thingToCloneID>`


Things can be printed thanks to a list command. 

Print a list of available things and their variables by using this command:

`$ arduino-cloud-cli thing list --show-variables`

Print a *filtered* list of available things, print only things belonging to the ids list:

`$ arduino-cloud-cli thing list --ids <thingOneID>,<thingTwoID>`

Print only the thing associated to the passed device:

`$ arduino-cloud-cli thing list --device-id <deviceID>`

Delete a thing with the following command:

`$ arduino-cloud-cli thing delete --device-id <deviceID>`

Extract a template from an existing thing. The template can be saved in two formats: json or yaml. The default format is yaml:

`$ arduino-cloud-cli thing extract --id <thingID> --outfile <templateFile> --format <yaml|json>`

Bind a thing to an existing device:

`$ arduino-cloud-cli thing bind --id <thingID> --device-id <deviceID>`

## Ota commands

Perform an OTA firmware update. Note that the binary file (`.bin`) should be compiled using an arduino core that supports the specified device.
The default OTA upload should complete in 10 minutes. Use `--deferred` flag to extend this time to one week.

`$ arduino-cloud-cli ota upload --device-id <deviceID> --file <sketch-file.ino.bin>`

`$ arduino-cloud-cli ota upload --device-id <deviceID> --file <sketch-file.ino.bin> --deferred`
