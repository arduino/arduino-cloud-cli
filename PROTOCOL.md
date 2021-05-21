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
  - request to know the connected thing and status; when this request is received, cloud will reply on topic /a/d/_deviceid_/shadow/i as described above


after initial connection

* Device can publish variables to a topic THING_OUT = /a/t/_thingid_/e/o
    
* Device can subscribe and receive variables changes from a topic THING_IN = /a/t/_thingid_/e/i

  
Connection types
---------------

* Arduino devices have a strong security level because at provisioning time the device certificate is written on a crypto-chip on the board itself. Hence, connection happens using that device certificate and thus confirming device identity
  
* Third party boards and other devices can connect to MQTT with a lower security level by using:
  - username = device_id    
  - password = device_secretkey
  
Both these information are provided during device configuration via API or on Cloud Web UI. Cloud API documentation is available here https://www.arduino.cc/reference/en/iot/api/   and at the same location you can find reference clients for Javascript, Python and Golang. 
  
  
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
  
  
