Arduino IoTCloud - MQTT protocol description
=======

on connection:

1. Device must subscribe to an incoming topic /a/d/<deviceid>/shadow/i to receive:
  - thingID that represents the Thing that this device is currently connected to and to which it should populate data
  - last status of all variables associated to this Thing
(which means: cloud is storing a Thing status reflecting values of certain variables; when device restarts,
it should align its internal status to the status last stored on cloud)

2. Device must post to an outgoing topic /a/d/<deviceid>/shadow/o to send:
  - any information about the device itself, like serial numbers, firmwware version, battery status
  - request to know the connected thing and status; when this request is received, cloud will reply on topic /a/d/<deviceid>/shadow/i as described above


after initial connection

* Device can publish collected data (properties) to a topic /a/t/<thingid>/e/o
