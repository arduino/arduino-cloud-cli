# arduino-cloud-cli

arduino-cloud-cli is a command line tool that empowers users to access the major features of Arduino IoT Cloud from a terminal. 

### License
This code is licensed under the terms of the GNU Affero General Public License v3.0. If you have questions about licensing or need a commercial license please contact us at [license@arduino.cc](mailto:license@arduino.cc).

### Requirements

This is all you need to use arduino-cloud-cli:
 * A client ID and a secret ID, retrievable from the [cloud](https://create.arduino.cc/iot/integrations) by creating a new API key

### Additional info

This tool follows a "quiet on success/verbose on error" behaviour. This means that when a command execution result is an error, such error is printed. On the other hand, when the command is successfully executed, there is no 'success' feedback on the output: the results of the command, if any, are directly printed without any other feedback. This strategy allows users to save the output of the command in a file. 

However, if the verbose flag `-v` is used, then the behaviour will change: the logs will always be printed out, providing users with feedback on the execution of the command. 

As an example, let's take the `device create` command. We want to save the information of the newly created device in a file.
So we simply lunch the command:

`$ arduino-cloud-cli device create --name mydevice --format json > mydevice.json`

The resulting mydevice.json will only contain device information in a json format.

Another example: let's say that the execution of the previous command results in an error. In that case, the json file will be empty and the terminal will print out the error. Now we want to execute the command again with the logs enabled in order to understand the issue. So we execute the following:

`$ arduino-cloud-cli device create --name mydevice -v`


## Set credentials 

arduino-cloud-cli needs a credentials file containing an Arduino IoT Cloud client ID and its corresponding secret.
Credentials can also include an optional organization ID that specifies the organization to use.
You can retrieve these credentials from the [cloud](https://create.arduino.cc/iot/integrations) by creating a new API key.

Once you have the credentials, execute the following command and provide them:

`$ arduino-cloud-cli credentials init`

By default it will be created in the Arduino data directory (Arduino15). 
You can specify a different destination folder with the `--dest-dir` option. 
arduino-cloud-cli looks for its credentials file in different directories in the following order: current working directory, parents of the current working directory, arduino15 default directory.

This gives you the possibility to use different credentials files depending on the project you are working on.

`$ arduino-cloud-cli credentials init --dest-dir <destinationFolder>`

To reset an old credentials file, just overwrite it using this command:

`$ arduino-cloud-cli credentials init --overwrite`

The credentials file is supported in two different formats: json and yaml. Use the `--file-format` to choose it. Default is yaml.

`$ arduino-cloud-cli credentials init --file-format json`

It is also possible to specify credentials directly in `ARDUINO_CLOUD_CLIENT`, `ARDUINO_CLOUD_SECRET` and optionally `ARDUINO_CLOUD_ORGANIZATION` environment variables. Credentials specified in environment variables have higher priority than the ones specified in credentials files.
Please note that credentials are correctly extracted from environment variables only if all the mandatory credentials parameters (client and secret) are found in environment variables. (think of it as another config file but with higher priority)

#### Find credentials

To have information about which credentials would be used in the current folder you can use the following command:

`$ arduino-cloud-cli credentials find`

## Device provisioning

When provisioning a device, you can optionally specify the port to which the device is connected and its fqbn. If they are not given, then the first device found will be provisioned.

Use this command to provision a device:

`$ arduino-cloud-cli device create --name <deviceName> --port <port> --fqbn <deviceFqbn>`

Here are the FQBNs of the Arduino boards that can be provisioned with this command:
```
- arduino:samd:nano_33_iot
- arduino:samd:mkrwifi1010
- arduino:mbed_nano:nanorp2040connect
- arduino:mbed_portenta:envie_m7
- arduino:mbed_nicla:nicla_vision
- arduino:samd:mkr1000
- arduino:samd:mkrgsm1400
- arduino:samd:mkrnb1500
```

#### LoRa 

LoRa devices should be provisioned using a specific command.
Parameters are the same except for the additional mandatory `--frequency-plan`:

`$ arduino-cloud-cli device create-lora --name <deviceName> --frequency-plan <freqID> --port <port> --fqbn <deviceFqbn>`

The list of supported LoRa frequency plans can be retrieved with:

`$ arduino-cloud-cli device list-frequency-plans`

Here are the FQBNs of the Arduino boards that can be provisioned with this command:
```
- arduino:samd:mkrwan1310
- arduino:samd:mkrwan1300
```

#### Generic device

A generic device is like a virtual device that doesn't need to be attached to an actual physical board.
Any actual physical board can connect to Arduino IoT Cloud using the credentials of a generic device.
Generic devices can be created using a specific command. 
An optional `--fqbn` flag can be passed to specify the fqbn of the device, otherwise it will be set to `generic:generic:generic`.

`$ arduino-cloud-cli device create-generic --name <deviceName> --fqbn <fqbn>`

The list of supported FQBN can be retrieved with:

`$ arduino-cloud-cli device list-fqbn`

## Device commands

Devices can be deleted using the device delete command. This command accepts two mutually exclusive flags: `--id` and `--tags`. Only one of them must be passed. When the `--id` is passed, the device having such ID gets deleted:

`$ arduino-cloud-cli device delete --id <deviceID>`

When `--tags` is passed, the devices having all the specified tags get deleted:

`$ arduino-cloud-cli device delete --tags <key0>=<value0>,<key1>=<value1>`

Devices currently present on Arduino IoT Cloud can be retrieved with:
`$ arduino-cloud-cli device list`

It has an optional `--tags` flag that allows to list only the devices having all the provided tags:
`$ arduino-cloud-cli device list --tags <key0>=<value0>,<key1>=<value1>`

Add tags to a device. Tags should be passed as a comma-separated list of `<key>=<value>` items:

`$ arduino-cloud-cli device create-tags --id <deviceID> --tags <key0>=<value0>,<key1>=<value1>`

Delete specific tags of a device. The keys of the tags to delete should be passed in a comma-separated list of strings:

`$ arduino-cloud-cli device delete-tags --id <deviceID> --keys <key0>,<key1>`

## Thing commands

Things can be created starting from a template or by cloning another thing.

Create a thing from a thing template. Supported template formats are JSON and YAML. The name parameter is optional. If it is provided, then it overrides the name retrieved from the template:

`$ arduino-cloud-cli thing create --name <thingName> --template <template.(json|yaml)>`

Create a thing by cloning another thing. Here the *name is mandatory*:

`$ arduino-cloud-cli thing clone --name <thingName> --clone-id <thingToCloneID>`


Things can be printed thanks to a list command. 

Print a list of available things and their variables by using this command:

`$ arduino-cloud-cli thing list --show-variables`

Print a *filtered* list of available things, print only things belonging to the ids list:

`$ arduino-cloud-cli thing list --ids <thingOneID>,<thingTwoID>`

Print only the thing associated to the passed device:

`$ arduino-cloud-cli thing list --device-id <deviceID>`

Print only the things that have all the passed tags:

`$ arduino-cloud-cli thing list --tags <key0>=<value0>,<key1>=<value1>`

Things can be deleted using the thing delete command. This command accepts two mutually exclusive flags: `--id` and `--tags`. Only one of them must be passed. When the `--id` is passed, the thing having such ID gets deleted:

`$ arduino-cloud-cli thing delete --id <thingID>`

When `--tags` is passed, the things having all the specified tags get deleted:

`$ arduino-cloud-cli thing delete --tags <key0>=<value0>,<key1>=<value1>`

Delete a thing with the following command:

`$ arduino-cloud-cli thing delete --id <thingID>`

Extract a template from an existing thing. The template is printed to stdout and its format depends on the global `--format` flag:

`$ arduino-cloud-cli thing extract --id <thingID> --format <json|yaml>`

Bind a thing to an existing device:

`$ arduino-cloud-cli thing bind --id <thingID> --device-id <deviceID>`

Add tags to a thing. Tags should be passed as a comma-separated list of `<key>=<value>` items:

`$ arduino-cloud-cli thing create-tags --id <thingID> --tags <key0>=<value0>,<key1>=<value1>`

Delete specific tags of a thing. The keys of the tags to delete should be passed in a comma-separated list of strings:

`$ arduino-cloud-cli thing delete-tags --id <thingID> --keys <key0>,<key1>`


## Ota commands

Perform an OTA firmware update. Note that the binary file (`.bin`) should be compiled using an arduino core that supports the specified device.
The default OTA upload should complete in 10 minutes. Use `--deferred` flag to extend this time to one week.

`$ arduino-cloud-cli ota upload --device-id <deviceID> --file <sketch-file.ino.bin>`

`$ arduino-cloud-cli ota upload --device-id <deviceID> --file <sketch-file.ino.bin> --deferred`

It is also possible to perform a mass ota upload through a specific command. 
The fqbn is mandatory. 
To select the devices to update you can either provide a list of device ids or device tags.

`$ arduino-cloud-cli ota mass-upload --fqbn <deviceFQBN> --device-ids <deviceIDs> --file <sketch-file.ino.bin>`

`$ arduino-cloud-cli ota mass-upload --fqbn <deviceFQBN> --device-tags <key0>=<value0>,<key1>=<value1> --file <sketch-file.ino.bin>`

## Dashboard commands

Print a list of available dashboards and their widgets by using this command:

`$ arduino-cloud-cli dashboard list --show-widgets`

Delete a dashboard with the following command:

`$ arduino-cloud-cli dashboard delete --id <dashboardID>`

Extract a template from an existing dashboard. The template is printed to stdout and its format depends on the global `--format` flag:

`$ arduino-cloud-cli dashboard extract --id <dashboardID> --format <json|yaml>`

Create a dashboard: dashboards can be created only starting from a template. Supported dashboard template formats are JSON and YAML. The name parameter is optional. If it is provided, then it overrides the name retrieved from the template. The `override` flag can be used to override the template `thing_id` placeholder with the actual ID of the thing to be used.

`$ arduino-cloud-cli dashboard create --name <dashboardName> --template <template.(json|yaml)> --override <thing-0>=<actualThingID>,<thing-1>=<otherActualThingID>`
