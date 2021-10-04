# Dashboard create command

## Review Period

Best before October, 5, 2021.

## Overview
Arduino Cloud CLI should allow users to create a dashboard starting from a cloud template.

The following User story has been taken from the [RFC](https://arduino.atlassian.net/wiki/spaces/FTWEB/pages/2761064740/Arduino+Cloud+CLI).
* User is able to launch a simple CLI command to create a Dashboard in Cloud
  * the new Dashboard to create is a logical copy of another existing “template” Dashboard
  * the newly created Dashboard is displaying variables coming from a Thing specified during creation

## Problem
**An implementation for this command should be chosen.**

The RFC specifies that this command is going to work only for dashboards with a single thing. However, dashboards with multiple things are common and **it would be better to introduce this feature from the beginning.**

The problem in supporting multiple things for a single dashboard lays in mapping the things described in the dashboard template to the actual things.
Indeed, if a single thing is used, it can be passed as a simple flag into the command.

## Accepted solution

It has been decided to implement the **Multiple things support** solution because it was the most future-proof.
### Multiple things support

Let's take a dashboard template:
```YAML
id: home-security-alarm-dashboard
name: Home Security Alarm
widgets:
  - type: Messenger
    name: message_update
    variables:
      - thing_id: home-security-messenger
        variable_id: message_update
  - type: Switch
    name: light_alarm
    variables:
      - thing_id: home-security-alarm
        variable_id: light_alarm
```
Here we are going to use a flag, that could be called `override`, which takes as parameter a map to be defined with the following structure: `<thing-placeholder>=<thing-id>,..`

Following the example, the command will be something like:

```sh
arduino-cloud-cli dashboard create --name <dashname> --template <yamldashtemplfile> --thing-override home-security-alarm=<mythingid1>,home-security-messenger=<mythingid2>
```


## Alternative solutions
### Single thing support

This is the simplest solution but it's very limiting.

Let's take a simple example:
```YAML
id: home-security-alarm-dashboard
name: Home Security Alarm
widgets:
  - type: Messenger
    name: message_update
    variables:
      - thing_id: home-security-alarm
        variable_id: message_update
```
In this example, the dashboard uses a single thing. So the command could easily be something like: 

```sh
arduino-cloud-cli dashboard create --name <dashname> --template <yamldashtemplfile> --thing-id <mythingid>
```

or even:

```sh
arduino-cloud-cli dashboard create --name <dashname> --template <yamldashtemplfile> --thing-override <home-security-alarm>=<mythingid>
```
