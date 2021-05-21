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


MQTT Topics
---------------

"IN" ("INPUT") topics are topics on which the device recives data. "OUT" ("OUTPUT") topics are topics on which the device publishes data

* THING_OUT = /a/t/_thingid_/e/o
* THING_IN = /a/t/_thingid_/e/i
* DEVICE_OUT = /a/d/_deviceid_/e/o
* DEVICE_SHADOW_IN = /a/d/_deviceid_/shadow/i
* DEVICE_SHADOW_OUT = /a/d/_deviceid_/shadow/o


Expected behavior on reset/connection:

1. Device MUST subscribe to DEVICE_SHADOW_IN  to receive:
  - thingID that represents the Thing that this device is currently associated to and to which it should populate data
  - last status of all variables associated to this Thing
when receiving a message on DEVICE_SHADOW_IN, the device MUST restore its internal status to the value of variables received, and properly set associated thingid

2. Device MUST publish to DEVICE_SHADOW_OUT to send:
  - RPC request of a method "getLastValues" to know the connected thing and status; when this request is received, cloud will reply on topic DEVICE_SHADOW_IN as described above

3. device can publish on DEVICE_OUT (optional) to send: 
  - any information about the device itself, like serial numbers, firmwware version, battery status. this will be stored as is by the cloud

after initial connection

* Device can publish variable changes to the topic THING_OUT   
* Device can subscribe and receive variable changes from a topic THING_IN  
* At any point in time, cloud can send a last status on DEVICE_SHADOW_IN to force a sync with local variables and also to communicate a change in thingid connected.
the change can also notify thingid = _UNASSOCIATED_ which means the device is not associated to a thing yet

  
Message format
---------------
  
* Messages are in SenML encoded in CBOR according to standard https://datatracker.ietf.org/doc/html/rfc8428 and https://datatracker.ietf.org/doc/html/rfc7049
  
Example of SenML encoding of properties
  
      [
        {"n":"temperature","v":120.1},
        {"n":"humidity","v":1.2},
        {"n":"fanstatus","vb":true}
      ]

Integers can be used for map keys as specified in RFC8428
  
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


  
