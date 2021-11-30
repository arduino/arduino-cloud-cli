#include <MKRWAN.h>
  
LoRaModem modem;
  
void setup() {
  Serial.begin(9600);
  while (!Serial);
  if (!modem.begin(EU868)) {
    Serial.println("Failed to start module");
    while (1) {}
  };
}

void loop() {
  Serial.println(modem.deviceEUI());
  delay(3000);
}
