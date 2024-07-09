#include <ArduinoIoTCloud.h>

#include "thingProperties.h"

void setup() {
  // Initialize serial and wait for port to open:
  Serial.begin(9600);
  // This delay gives the chance to wait for a Serial Monitor without blocking if none is found
  delay(1500); 

  // Defined in thingProperties.h
  initProperties();

  // Connect to Arduino IoT Cloud
  ArduinoCloud.begin(ArduinoIoTPreferredConnection, false, "mqtts-sa.iot.oniudra.cc");

  setDebugMessageLevel(4);
  ArduinoCloud.printDebugInfo();
}

unsigned long previousMillis = 0;
const long interval = 5000; //ms
bool increase = true;

// the loop function runs over and over again forever
void loop() {
  ArduinoCloud.update();
  
  unsigned long currentMillis = millis();
  if (currentMillis - previousMillis >= interval) {
    previousMillis = currentMillis;
    increase = !increase;

    if(pressure < 2){
      pressure = 8;
    }
    if(temperature < 2){
      temperature = 25;
    }

    int randNumber = random(10.0, 70.0);
    if(!increase){
      randNumber = -randNumber;
    }
    float diff = (float)randNumber / 100.0;

    pressure = pressure + diff;
    temperature = temperature + diff;

    Serial.println(AIOT_CONFIG_LIB_VERSION);
    Serial.println("2.1");
  }
  
}

/*
  Since Temperature is READ_WRITE variable, onTemperatureChange() is
  executed every time a new value is received from IoT Cloud.
*/
void onTemperatureChange()  {
  // Add your code here to act upon Temperature change
}

/*
  Since Pressure is READ_WRITE variable, onPressureChange() is
  executed every time a new value is received from IoT Cloud.
*/
void onPressureChange()  {
  // Add your code here to act upon Pressure change
}
