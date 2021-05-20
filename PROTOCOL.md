Arduino IoTCloud - MQTT protocol description
=======

on connection:

1. Device must subscribe to an incoming topic DEVICE_SHADOW_IN = /a/d/_deviceid_/shadow/i to receive:
  - thingID that represents the Thing that this device is currently connected to and to which it should populate data
  - last status of all variables associated to this Thing
(which means: cloud is storing a Thing status reflecting values of certain variables; when device restarts,
it should align its internal status to the status last stored on cloud)

2. Device must post to an outgoing topic DEVICE_SHADOW_OUT = /a/d/_deviceid_/shadow/o to send:
  - any information about the device itself, like serial numbers, firmwware version, battery status
  - request to know the connected thing and status; when this request is received, cloud will reply on topic /a/d/<deviceid>/shadow/i as described above


after initial connection

* Device can publish variables to a topic THING_OUT = /a/t/_thingid_/e/o
    
* Device can subscribe and receive variables from a topic THING_IN = /a/t/_thingid_/e/i

  
  
Message format
---------------
  
* Messages are in SenML encoded in CBOR according to standard https://datatracker.ietf.org/doc/html/rfc8428 and https://datatracker.ietf.org/doc/html/rfc7049

example of SenML coding of properties
  
      [
        {"n":"temperature","v":120.1},
        {"n":"humidity","v":1.2},
        {"n":"fanstatus","vb":true}
      ]
  
Note: use http://cbor.me/  to easily get a CBOR representation starting from JSON equivalent. also see http://cbor.io/
  
Example
---------------
  

  
  
