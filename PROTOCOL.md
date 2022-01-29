Arduino IoTCloud - MQTT protocol description
=======


Core concepts
--------------

* Device is a physical IoT device identified by a deviceid
* Thing is a logical representation of a device in Arduino Cloud, also known as "Digital Twin" for the Device. A Thing has its own thingid. 
* A Device has a 1-to-1 association with a Thing. However, there is a phase of the lifecycle in which a device is connected but the corresponding thing might still not exist. Plus, a device can be detached from a Thing and associated to another Thing. In a similar way, for example when a device breaks, the corresponding Thing in Cloud still remains and can be associated to another Device. In this way, the historical status and information of this Thing are preserved even in case of hardware failures of the Device
* MQTT broker is the frontend communication server that communicates with the device. Coimmunication happens via messages on MQTT topics.
* a Thing has a status that is based on a set of variables that can change over time. For example, the status of a switch or a temperature
* the Thing status is persisted on cloud and also historical changes are tracked
* a Device has a local copy of the status of its associated Thing; however, this local copy can be lost due to a failure or reset of the device. For this reason, the cloud can use _SHADOW_ topics to restore the proper status of a device making it identical to the remote copy. Let's consider an example: a device has a variable to control an engine speed, with curent value 30 rpm. The device has a restart and the variable goes to a default value of 0 rpm. However, immediately after restart, the connected thing will send information to the device (using SHADOW topic) to restore the last value of 30 rpm in its local state. Hence, the control continues to work as expected.


 
                                                          °°°°°°°°°
                                                  °°°°°°°°°°°°  °°°°°°
                                              °°°  °°         °    °°° °°°
      ┌─────────┐                            °°°                 ┌──────┐     °
      │         │                         ┌─────────────┐        │      │     °
      │ device  ├────────────────────────▲│MQTT broker  ├───────▲│thing │   °°
      │         │▼────────────────────────┴─────────────┘▼───────┤      │   ° °°
      └─────────┘                           °                    └──────┘      °
                                            ° °           °         °          °
                                               ° °°      °°°        °°°° ° ° °°
                                                   ° °° °  °°° ° ° °


### MQTT Topics

"IN" ("INPUT") topics are topics on which the device recives data. "OUT" ("OUTPUT") topics are topics on which the device publishes data

* THING_OUT = "/a/t/_thingid_/e/o"    this topic is used to notify all clients that there was a change in Thing status 
* THING_IN = "/a/t/_thingid_/e/i"   this topic is used by clients that want to request a change to a Thing status      

* DEVICE_OUT = "/a/d/_deviceid_/e/o"    this topic is used by the device to publish device config info to cloud
* DEVICE_IN = "/a/d/_deviceid_/e/i"    this topic is used by the device to receive device config change requests from cloud
      
* THING_SHADOW_OUT = /a/d/_thingid_/shadow/o    is used by the device connected to this thing to request last thing status to cloud
* THING_SHADOW_IN = /a/d/_thingid_/shadow/i     used by the cloud to communicate last thing status to the connected device



### Scenario 1: RESET

1. Immediately after start, the Device publishes its current configuration to cloud on DEVICE_OUT topic
Device configuration is a set of properties related to device capability, current firmware version,...
for example the initial device config message can be

 #device
 PUBLISH /a/d/DEVICE_ID/e/o
 {
     "OTA_CAP": true,
     "OTA_ERROR": 0,
     "OTA_SHA256": "73475cb40a568e8da8a045ced110137e159f890ac4da883b6b17dc651b3a8049"
 }


2. Device subscribes to DEVICE_IN  to receive a configuration update from cloud.
Most important configuration update is the "thingID" configuration, that represents the Thing that this device is currently associated to and to which it should populate data. The device will wait until the config update request is received. If nothing is received after a timeout, the device shall unsubscribe and subscribe again to this topic to trigger a new configuration update request

 #device
 SUBSCRIBE /a/d/DEVICE_ID/e/i

3. Cloud sends a configuration update on DEVICE_IN topic 
this is triggered by the fact that the device subscribed to DEVICE_IN topic. every time the device subscribes, it will receive device config properties

 #cloud
 PUBLISH /a/d/DEVICE_ID/e/i
 {
     "thing_id": "e505ab27-01b5-43a3-9119-d5e9bcd3f1d1"
 }
  
Note that cloud might send an empty string for ThingID, to signal that the device is currently not attached to a thing. in such case, the device is not authorized to send data, might report a warning to the user.

4. at this point, the device knows the connected thing, hence it can subscribe to proper topics to receive input data (THING_IN and THING_SHADOW_IN)

 #device
 SUBSCRIBE /a/t/THING_ID/e/i
 SUBSCRIBE /a/t/THING_ID/shadow/i

5. in order to restore the last status of the thing on device , the device performs an RPC request of a method "getLastValues" to know the connected thing status 

 #device
 PUBLISH /a/t/THING_ID/shadow/o
 {
     "r:m": "getLastValues"
 }

6. cloud replies to the getLastValues request sending last thing status (all variables) on THING_SHADOW_IN

 #cloud
 PUBLISH /a/t/THING_ID/shadow/i
 {
     "temperature": 27,
     "humidity": 0.5
 }

7. the device will now mirror the status change to THING_OUT 

 #device
 PUBLISH /a/t/THING_ID/e/o
 {
     "temperature": 27,
     "humidity": 0.5
 }

Note: not sure why this is needed, the thing status is already as such...


### Scenario 2: thing status updates 

* Device can publish variable changes to the topic THING_OUT

 #device
 PUBLISH /a/t/THING_ID/e/o
 {
     "temperature": 29,
     "humidity": 0.5 
 }


* Device can receive variable changes from the topic THING_IN  

 #cloud
 PUBLISH /a/t/THING_ID/e/i
 {
     "temperature": 29,
     "humidity": 0.5 
 }


Note: at the moment, the expectation is that when a variable change is received on THING_IN topic, the device will apply the variable change and then mirror it back on the THING_OUT topic so that all other listeners will realize that there was a change. One of the listeners is also the cloud which will save this status change in the thing status.
For example: when a dashboard widget is used to set a variable switch=true, this change request is published by the widget on THING_IN, then the device is applying it and publishing on THING_OUT so that all other listeners (for example other widgets on other dashboards that is not the one which changed the variable) can also be made aware. 
However, this behavior is questionable because it requires the device to be online. ---- TO BE FURTHER DISCUSSED ----
Proposed behavior (to be discussed): 
* widget requests variable change on THING_IN
* it is cloud responsibility to apply the change to thing_status and mirror the changed status on THING_OUT for all listeners 
* the device is still registered to THING_IN and will still apply the change but doesn't have responsibility to mirror on THING_OUT
* if the device is offline, the change to thing status happens anyway, and when the device returns online, it will be informed about the change via getLastValues


### Scenario 3: OTA update

this is a valid scenario only for devices that have sent the OTA_CAP variable as true, i.e. they support OTA update

1. cloud can PREPARE an OTA update request by sending on THING_SHADOW_IN OTA_URL with the url of OTA file to download and apply 

 #cloud
 PUBLISH /a/t/THING_ID/shadow/i
 {
     "OTA_URL": "https://api2.arduino.cc/iot/ota/9ef3ec89-b4e4-4dc2-8f24-c5285a6e768c",
     "OTA_REQ": false 
     ... there might be other props...
 }

note that this is still only communicating OTA_URL while for preparing the device, it's not requesting the OTA update to start

2. device will mirror back the OTA_URL as part of the mirror of normal variables

 #device
 PUBLISH /a/t/THING_ID/e/o
 {
     "OTA_URL": "https://api2.arduino.cc/iot/ota/9ef3ec89-b4e4-4dc2-8f24-c5285a6e768c",
     "OTA_REQ": false 
     ... there might be other props...
 }

3. now cloud will publish a message on THING_SHADOW_IN with OTA_REQ=true to START the OTA update
This is a request to the device to start the OTA update process, download the OTA file from OTA_URL, apply and restart.

 #cloud
 PUBLISH /a/t/THING_ID/shadow/i
 {
     "OTA_URL": "https://api2.arduino.cc/iot/ota/9ef3ec89-b4e4-4dc2-8f24-c5285a6e768c",
     "OTA_REQ": true 
     ... there might be other props...
 }

4. the device is executing OTA download, and restart
after restart (see scenario 1) the first message is the device sending its current configuration, which contains
variables describing OTA result and in particular:
* OTA_ERROR = 0 if no error, 1 if there was an error, and the device had to rollback and use previous firmware
* OTA_SHA256 = the SHA256 of currently running firmware (it could be the last OTA FW or the previous one if OTA failed)

 #device
 PUBLISH /a/d/DEVICE_ID/e/o
 {
     "OTA_CAP": true,
     "OTA_ERROR": 0,
     "OTA_SHA256": "73475cb40a568e8da8a045ced110137e159f890ac4da883b6b17dc651b3a8049"
 }

cloud can use these information to understand if OTA was successful or not

Note/Question: is OTA_SHA256 having a proper value also for local uploads ? or is it only for OTA because it's in the OTA file ? 
what is the device config message like if there was never an OTA executed ? 

  
Message format
---------------
  
* Messages are in SenML encoded in CBOR according to standard https://datatracker.ietf.org/doc/html/rfc8428 and https://datatracker.ietf.org/doc/html/rfc7049
  
Example of SenML encoding of properties
  
      [
        {"n":"temperature","v":120.1},
        {"n":"humidity","v":1.2},
        {"n":"fanstatus","vb":true}
      ]

Integers are used for map keys as specified in RFC8428
  
                  +---------------+-------+------------+
                  |          Name | Label | CBOR Label |
                  +---------------+-------+------------+
                  |  Base Version | bver  |         -1 |
                  |     Base Name | bn    |         -2 |
                  |     Base Time | bt    |         -3 |
                  |     Base Unit | bu    |         -4 |
                  |    Base Value | bv    |         -5 |
                  |      Base Sum | bs    |         -6 |
                  |          Name | n     |          0 |
                  |          Unit | u     |          1 |
                  |         Value | v     |          2 |
                  |  String Value | vs    |          3 |
                  | Boolean Value | vb    |          4 |
                  |           Sum | s     |          5 |
                  |          Time | t     |          6 |
                  |   Update Time | ut    |          7 |
                  |    Data Value | vd    |          8 |
                  +---------------+-------+------------+
  
  
Note: use http://cbor.me/  to easily get a CBOR representation starting from JSON equivalent. also see http://cbor.io/

* Arduino Cloud supports variable types that are in the form of a *Multi-Value map* with multiple sub-keys; those sub-keys are sent as separate variables in SenML, with the notation *variable:subkey*. For example, _DimmedLight_ is a variable type that has two sub-keys "swi" of type bool which represents the on/off status of the light, and "bri" of type integer which represents the light brightness from 0 to 100. Hence if you define a variable _dimlit_ of type _DimmedLight_, the SenML record will be

       [{0: "dimlit:swi", 4: false}, {0: "dimlit:bri", 2: 59}]

* *Methods* can be invoked to the cloud as in a form of RPC. To invoke a method, a special property name "r:m" is used, while the property value is the method name
  
  
  
Example
---------------
  
 - device writes following message to DEVICE_SHADOW_OUT
                     
          [{0: "r:m", 3: "getLastValues"}]   => in CBOR =>   81 A2 00 63 72 3A 6D 03 6D 67 65 74 4C 61 73 74 56 61 6C 75 65 73
  
  - response from cloud on DEVICE_SHADOW_IN  ( device shall set its internal status to these values ) 
          
          [{0: "temperature", 2: 27.8}]
  
  - device sends current value of properties on THING_OUT

          [{0: "temperature", 2: 28.8}]
  
  
  
Connection types
---------------

* Arduino devices have a strong security level because at provisioning time the device certificate is written on a crypto-chip on the board itself. Hence, connection happens using that device certificate and thus confirming device identity
  
* Third party boards and other devices can connect to MQTT with a lower security level by using:
  - username = device_id    
  - password = device_secretkey
  
Both these information are provided during device configuration via API or on Cloud Web UI. Cloud API documentation is available here https://www.arduino.cc/reference/en/iot/api/   and at the same location you can find reference clients for Javascript, Python and Golang. 


  
