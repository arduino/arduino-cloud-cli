# Title: Thing Discovery Protocol 1.0

Use this as a template to write a design document when adding new
major features to the project. It helps other developers
understand the scope of the project, validate technical complexity and
feasibility. It also serves as a public documentation of how the feature
actually works.

This document describes the workflow to allow a board to get a proper configuraition from the cloud.

## Names and Conventions

 - Properties: is the old name, Variable is the new one, but internally the platform is still using the name "property".
 - Board: phisical device running the sketch.
 - Cloud: the Arduino IoT Cloud platform as a whole.

## Review Period

Best before Nov, 26th 2021.

## What is the problem?

The current version of the IoT Cloud Library expects the sketch to contain the Thing ID as a hardcoded string.

Having the Thing ID hard coded in the sketch binary limits the reuse of the sketch on different boards.


### Constraints

Required information:
 - Device ID

### Workflow (current)

Here's the current workflow.

#### 1. Board: Connecto to the MQTT Broker

```
# board
CONNECT broker.arduino.cc
```

#### 2. Board: Subscribe to the Input Topics

```
# board
SUBSCRIBE /a/t/THING_ID/e/i
SUBSCRIBE /a/t/THING_ID/shadow/i
```

#### 3. Board: Send a "getLastValue" Request to the Cloud

```
# board
PUBLISH /a/t/THING_ID/shadow/o
{
    "r:m": "getLastValues"
}
```

#### 4. Cloud: Send a "getLastValue" Response to the Board

messages from cloud to board are not limited in size (as big as they can be)

```
# cloud
PUBLISH /a/t/THING_ID/shadow/i
{
    "temperature": 27,
    "humidity": 0.5,
    "OTA_URL": "https://pornhub.com",
    "OTA_REQ": false,
    "tz_offset": 0,
    "tz_dst_until": 0,
}
```

From board to cloud the message size is limited to 256 bytes.

```
# board
/a/t/THING_ID/e/o
{
    "temperature": 27,
    "humidity": 0.5,
    "OTA_URL": "https://api2.arduino.cc/iot/ota/9ef3ec89-b4e4-4dc2-8f24-c5285a6e768c",
    "OTA_REQ": false,
    "OTA_SHA256": "122221",
    "OTA_CAP": true,
    "OTA_ERROR": 0,
    "tz_offset": 0,
    "tz_dst_until": 0,
}
```

When the board receives the last values it can start the continue running the main loop serving and receiving variable changes to and from the the cloud.

This specific sketch binary will only work on one specific board and cannot be reused across two or more boards.

## Out-of-Scope

 - Dynamic board recofiguration.


## User Experience Walkthrough

Here's an update workflow that requires the board to know its device ID only.

### Constraints

Required information:
 - Device ID

### Workflow (updated)

#### 1. Board: Connecto to the MQTT Broker

```
# board
CONNECT broker.arduino.cc
```

#### 2. Board: Send current device properties to the

```
# board
PUBLISH /a/d/DEVICE_ID/e/o
{
    "OTA_CAP": true,
    "OTA_ERROR": 0,
    "OTA_SHA256": "1234567"
}
```

#### 3. Board: Subscribe to the device input topic

```
SUBSCRIBE /a/d/DEVICE_ID/e/i

#  —— waiting —— (UNSUBSCRIBE and re-SUBSCRIBE if messages are not received)
```

#### 4. Cloud: send the device configuration

messages from cloud to board are not limited in size (as big as they can be)

```
# cloud
PUBLISH /a/d/DEVICE_ID/e/i
{
    "thing_id": "123",
    "OTA_URL": "pornhub.com",
    "OTA_REQ": false
}
```


#### 5. Board: subscribe to the thing topic

```
SUBSCRIBE /a/t/THING_ID/shadow/i
```

#### 6. Cloud: publish the last values for the thing

```
# board
PUBLISH /a/t/THING_ID/shadow/i
{
    "temperature": 27,
    "humidity": 0.5,
    "OTA_URL": "https://pornhub.com",
    "OTA_REQ": false,
    "tz_offset": 0,
    "tz_dst_until": 0,
}
```

From board to cloud the message size is limited to 256 bytes.

```
# board
PUBLISH /a/t/THING_ID/e/o
{
    "temperature": 27,
    "humidity": 0.5,
    "OTA_URL": "https://pornhub.com",
    "OTA_REQ": false,
    "OTA_SHA256": "122221",
    "OTA_CAP": true,
    "OTA_ERROR": 0,
    "tz_offset": 0,
    "tz_dst_until": 0,
}
```


## Implementation

### Project Changes

_Explain the changes to api or command line interface, including adding new
commands, modifying arguments etc_

### Breaking Change

_Are there any breaking changes to the interface? Explain_

### Design

_Explain how this feature will be implemented. Highlight the components
of your implementation, relationships_ _between components, constraints,
etc._

### Security

_Tip: How does this change impact security? Answer the following
questions to help answer this question better:_

**What new dependencies (libraries/cli) does this change require?**

**What other Docker container images are you using?**

**Are you creating a new HTTP endpoint? If so explain how it will be
created & used**

**Are you connecting to a remote API? If so explain how is this
connection secured**

**Are you reading/writing to a temporary folder? If so, what is this
used for and when do you clean up?**

### Documentation Changes

_Explain the changes required to internal and public documentation (API reference, tutorial, etc)_

## Additional Notes

_Link any useful metadata: Jira task, GitHub issue, …_
