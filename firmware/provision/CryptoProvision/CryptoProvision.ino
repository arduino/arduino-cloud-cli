/*
   This file is part of ArduinoIoTCloud.

   Copyright 2019 ARDUINO SA (http://www.arduino.cc/)

   This software is released under the GNU General Public License version 3,
   which covers the main part of arduino-cli.
   The terms of this license can be found at:
   https://www.gnu.org/licenses/gpl-3.0.en.html

   You can be released from the requirements of the above licenses by purchasing
   a commercial license. Buying such a license is mandatory if you want to modify or
   otherwise use the software for commercial activities involving the Arduino
   software without disclosing the source code of your own applications. To purchase
   a commercial license, send an email to license@arduino.cc.
*/

#include <ArduinoIoTCloud.h>
#include "ECCX08TLSConfig.h"

#include "uCRC16Lib.h"
const uint8_t SKETCH_INFO[] = {0x55, 0xaa, 0x01, 0x00, 0x01, 0xff, 0xaa, 0x55};
const bool DEBUG = true;

ArduinoIoTCloudCertClass Cert;
CryptoUtil Crypto;

enum class MESSAGE_TYPE { NONE = 0, COMMAND, DATA, RESPONSE };
enum class COMMAND {
  GET_SKETCH_INFO = 1,
  GET_CSR,
  SET_LOCKED,
  GET_LOCKED,
  WRITE_CRYPTO,
  BEGIN_STORAGE,
  SET_DEVICE_ID,
  SET_YEAR,
  SET_MONTH,
  SET_DAY,
  SET_HOUR,
  SET_VALIDITY,
  SET_CERT_SERIAL,
  SET_AUTH_KEY,
  SET_SIGNATURE,
  END_STORAGE,
  RECONSTRUCT_CERT

};
enum class PROVISIONING_ERROR : uint8_t {
  NONE = 0,
  SYNC,
  LOCK_FAIL,
  LOCK_SUCCESS,
  WRITE_CONFIG_FAIL,
  CRC_FAIL,
  CSR_GEN_FAIL,
  CSR_GEN_SUCCESS,
  SKETCH_UNKNOWN,
  GENERIC,
  NO_DATA
};

enum class RESPONSE {
  RESPONSE_NONE = 1,
  RESPONSE_ACK,
  RESPONSE_NACK,
  RESPONSE_ERROR
};

#define MAX_PAYLOAD_LENGTH 130
#define CRC_SIZE 2
uint8_t
payloadBuffer[MAX_PAYLOAD_LENGTH +
              CRC_SIZE]; // max 64 bytes will be stored before next round
uint8_t msgStart[] = {0x55, 0xaa};
uint8_t msgEnd[] = {0xaa, 0x55};
MESSAGE_TYPE msgType = MESSAGE_TYPE::NONE;
uint16_t msgLength = 0;
uint16_t msgByteIndex = 0;
// message is structured as such {START H}{START L}{TYPE}{LENGTH H}{LENGHT L}{PAYLOAD}{PAYLOAD CRC H}{PAYLOAD CRC L}{END}
// minimum length is by commands with no payload 2+1+1+1+1+2+2 => 10

const uint16_t minMessageLength = 10;

enum class MACHINE_STATE {
  IDLE = 0,
  RECEIVING_PAYLOAD,
  PROCESS_CMD,
  PROCESS_MSG,
  PROCESS_MSG_END,
  SENDING,
  WRITING,
  LOCKING
};


MACHINE_STATE machineState = MACHINE_STATE::IDLE;
uint8_t deviceIDBytes[72];
String deviceIDstring;
String csr;

String issueYear;
String issueMonth;
String issueDay;
String issueHour;
String expireYears;
String authorityKeyIdentifier;
String signature;

void setup() {

  Serial.begin(57600);
  Serial1.begin(115200);
  uint8_t cryptoInitOK = cryptoInit();
  if (!cryptoInitOK) {
    Serial1.println("FAILED TO DETECT CRYPTO ELEMENT");
    while (1);
  } else {
    Serial1.println("Crypto-element initialized");
  }
  if (cryptoLock() != PROVISIONING_ERROR::LOCK_SUCCESS) {
    Serial1.println("FAILED TO LOCK CRYPTO ELEMENT");
    while (1);
  } else {
    Serial1.println("Crypto-element locked");
  }
}

void loop() {
  if (machineState == MACHINE_STATE::IDLE) {
    waitForMessage();

  }
  if (machineState == MACHINE_STATE::RECEIVING_PAYLOAD) {
    payloadBuffer[msgByteIndex] = (uint8_t)Serial.read();
    Serial1.print(payloadBuffer[msgByteIndex], HEX);
    Serial1.print(" ");
    msgByteIndex++;
    if (msgByteIndex >= msgLength) {
      changeState(MACHINE_STATE::PROCESS_MSG_END);
    }
  }
  if (machineState == MACHINE_STATE::PROCESS_MSG_END) {
    checkMessageEnd();
  }

}

void waitForMessage() {
  if (Serial.available() >= minMessageLength) {
    uint8_t msgStartBuffer[2];
    uint8_t byteIn;
    bool msgStartByteOK = false;
    while (!msgStartByteOK && Serial.available()) {
      byteIn = (uint8_t)Serial.read();
      if (byteIn == msgStart[0]) {
        msgStartBuffer[0] = byteIn;
        byteIn = (uint8_t)Serial.read();
        if (byteIn == msgStart[1]) {
          msgStartBuffer[1] = byteIn;
          msgStartByteOK = true;
        }
      }
    }

    //Serial.readBytes(msgStartBuffer, sizeof(msgStart));
    if (memcmp(msgStartBuffer, msgStart, sizeof(msgStart)) == 0) {
      Serial1.println("message START");
      msgType = (MESSAGE_TYPE)Serial.read();
      uint8_t lengthH = (uint8_t)Serial.read();
      uint8_t lengthL = (uint8_t)Serial.read();
      Serial1.print(lengthH);
      Serial1.print(" - ");
      Serial1.println(lengthL);
      msgLength = lengthH << 8 | lengthL;

      Serial1.print("TYPE: ");
      Serial1.println((int)msgType);
      Serial1.print("LENGTH: ");
      Serial1.println((int)msgLength);

      //delay(1000);

      if (msgLength > 0) {
        changeState(MACHINE_STATE::RECEIVING_PAYLOAD);
      } else {
        changeState(MACHINE_STATE::PROCESS_MSG_END);
      }
    }
  }
}
void checkMessageEnd() {
  if (Serial.available() >= sizeof(msgEnd)) {
    uint8_t msgEndBuffer[2];
    Serial.readBytes((char*)msgEndBuffer, sizeof(msgEnd));
    if (memcmp(msgEndBuffer, msgEnd, sizeof(msgEnd)) == 0) {
      Serial1.println("message END");
      if (processMessage() == PROVISIONING_ERROR::CRC_FAIL) {
        Serial1.println("ERROR:: CRC FAIL");
        sendData(MESSAGE_TYPE::RESPONSE, (char*)RESPONSE::RESPONSE_NACK, 1);
      }
      //delay(2000);
      // Serial.end();
      // for(;;) {
      //   delay(500);
      // };
      changeState(MACHINE_STATE::IDLE);
    }
  }
}

PROVISIONING_ERROR processMessage() {
  bool checkSumOK = false;
  if (msgLength > 0) {
    // checksum verification
    // uint8_t csHI = payloadBuffer[msgLength - 2];
    // uint8_t csLO = payloadBuffer[msgLength - 1];
    // char receivedCS[] = {csHI, csLO};
    uint16_t receivedCRC = ((uint16_t)payloadBuffer[msgLength - 2] << 8 | payloadBuffer[msgLength - 1]);
    uint16_t computedCRC = uCRC16Lib::calculate((char *)payloadBuffer, msgLength - CRC_SIZE);
    Serial1.print("DATA CRC: ");
    Serial1.println(receivedCRC, HEX);

    Serial1.print("COMPUTED CRC: ");
    Serial1.println(computedCRC, HEX);
    if (receivedCRC != computedCRC) return PROVISIONING_ERROR::CRC_FAIL;
    Serial1.println("CRC aligned");
    checkSumOK = true;
  }

  if (msgType == MESSAGE_TYPE::COMMAND) {
    processCommand();
  }
  if (msgType == MESSAGE_TYPE::DATA) {
    processRawData(checkSumOK);
  }
  return PROVISIONING_ERROR::NONE;
}

void processCommand() {
  Serial1.print("%%%%% ");
  Serial1.println(">> processing command");
  COMMAND cmdCode = (COMMAND)payloadBuffer[0];
  if (cmdCode == COMMAND::GET_SKETCH_INFO) {
    Serial1.println("get sketch info");
    char response[] = {char(RESPONSE::RESPONSE_ACK)};
    sendData(MESSAGE_TYPE::RESPONSE, response, 1);
  }

  if (cmdCode == COMMAND::GET_CSR) {
    // extract payload from [1] to [payloadLength]
    // this will be the device_id used to generate a valid CSR
    Serial1.println("get CSR");
    for (uint8_t i = 1; i < msgLength - CRC_SIZE; i++) {
      deviceIDBytes[i - 1] = payloadBuffer[i];
    }

    // clear device ID string
    // this will be sent to the host
    deviceIDstring = "";
    Serial1.print("Device ID from host: ");
    char charBuffer[2];
    for (uint8_t i = 0; i < msgLength - CRC_SIZE - 1; i++) {
      Serial1.print(deviceIDBytes[i], HEX);
    }

    deviceIDstring = String((char*)deviceIDBytes);

    Serial1.println();
    Serial1.print("request for CSR with device ID ");
    Serial1.println(deviceIDstring);

    if (generateCSR() == PROVISIONING_ERROR::CSR_GEN_SUCCESS) {
      sendData(MESSAGE_TYPE::DATA, csr.c_str(), csr.length());
      Serial1.println("CSR GENERATED ON BOARD");
    } else {
      Serial1.println("SOMETHING WENT WRONG");
      while (1);
    }
  }
  if (cmdCode == COMMAND::BEGIN_STORAGE) {
    Serial1.println("begin storage");
    if (!Crypto.writeDeviceId(deviceIDstring, CryptoSlot::DeviceId)) {
      Serial1.println("Error storing device id!");
      char response[] = {char(RESPONSE::RESPONSE_ERROR)};
      sendData(MESSAGE_TYPE::RESPONSE, response, 1);
      return;
    }
    if (!Cert.begin()) {
      Serial1.println("Error starting Crypto storage!");
      char response[] = {char(RESPONSE::RESPONSE_ERROR)};
      sendData(MESSAGE_TYPE::RESPONSE, response, 1);
      return;
    }
    Cert.setSubjectCommonName(deviceIDstring);
    Cert.setIssuerCountryName("US");
    Cert.setIssuerOrganizationName("Arduino LLC US");
    Cert.setIssuerOrganizationalUnitName("IT");
    Cert.setIssuerCommonName("Arduino");
    char response[] = {char(RESPONSE::RESPONSE_ACK)};
    sendData(MESSAGE_TYPE::RESPONSE, response, 1);
  }


  if (cmdCode == COMMAND::SET_YEAR) {
    Serial1.println("set year");
    char yearBytes[4];
    String yearString;
    for (uint8_t i = 1; i < msgLength - CRC_SIZE; i++) {
      yearBytes[i - 1] = payloadBuffer[i];
    }
    Serial1.print("Year from host: ");
    char charBuffer[2];
    for (uint8_t i = 0; i < msgLength - CRC_SIZE - 1; i++) {
      Serial1.print(yearBytes[i], HEX);
      sprintf(charBuffer, "%d", yearBytes[i]);
      yearString += String(yearBytes[i]);//String(deviceIDBytes[i], 16);
    }

    Serial1.println();
    Serial1.print("set Cert YEAR to ");
    Serial1.println(yearString);
    Cert.setIssueYear(yearString.toInt());

    char response[] = {char(RESPONSE::RESPONSE_ACK)};
    sendData(MESSAGE_TYPE::RESPONSE, response, 1);

  }
  if (cmdCode == COMMAND::SET_MONTH) {
    Serial1.println("set month");
    char monthBytes[4];
    String monthString;
    for (uint8_t i = 1; i < msgLength - CRC_SIZE; i++) {
      monthBytes[i - 1] = payloadBuffer[i];
    }
    Serial1.print("month from host: ");
    char charBuffer[2];
    for (uint8_t i = 0; i < msgLength - CRC_SIZE - 1; i++) {
      Serial1.print(monthBytes[i], HEX);
      sprintf(charBuffer, "%d", monthBytes[i]);
      monthString += String(monthBytes[i]);//String(deviceIDBytes[i], 16);
    }

    Serial1.println();
    Serial1.print("set Cert MONTH to ");
    Serial1.println(monthString);
    Cert.setIssueMonth(monthString.toInt());

    char response[] = {char(RESPONSE::RESPONSE_ACK)};
    sendData(MESSAGE_TYPE::RESPONSE, response, 1);

  }

  if (cmdCode == COMMAND::SET_DAY) {
    Serial1.println("set day");
    char dayBytes[4];
    String dayString;
    for (uint8_t i = 1; i < msgLength - CRC_SIZE; i++) {
      dayBytes[i - 1] = payloadBuffer[i];
    }
    Serial1.print("day from host: ");
    char charBuffer[2];
    for (uint8_t i = 0; i < msgLength - CRC_SIZE - 1; i++) {
      Serial1.print(dayBytes[i], HEX);
      sprintf(charBuffer, "%d", dayBytes[i]);
      dayString += String(dayBytes[i]);//String(deviceIDBytes[i], 16);
    }

    Serial1.println();
    Serial1.print("set Cert day to ");
    Serial1.println(dayString);
    Cert.setIssueDay(dayString.toInt());

    char response[] = {char(RESPONSE::RESPONSE_ACK)};
    sendData(MESSAGE_TYPE::RESPONSE, response, 1);

  }

  if (cmdCode == COMMAND::SET_HOUR) {
    Serial1.println("set hour");
    char hourBytes[4];
    String hourString;
    for (uint8_t i = 1; i < msgLength - CRC_SIZE; i++) {
      hourBytes[i - 1] = payloadBuffer[i];
    }
    Serial1.print("hour from host: ");
    char charBuffer[2];
    for (uint8_t i = 0; i < msgLength - CRC_SIZE - 1; i++) {
      Serial1.print(hourBytes[i], HEX);
      sprintf(charBuffer, "%d", hourBytes[i]);
      hourString += String(hourBytes[i]);//String(deviceIDBytes[i], 16);
    }

    Serial1.println();
    Serial1.print("set Cert hour to ");
    Serial1.println(hourString);
    Cert.setIssueHour(hourString.toInt());

    char response[] = {char(RESPONSE::RESPONSE_ACK)};
    sendData(MESSAGE_TYPE::RESPONSE, response, 1);

  }

  if (cmdCode == COMMAND::SET_VALIDITY) {
    Serial1.println("set validity");
    char validityBytes[4];
    String validityString;
    for (uint8_t i = 1; i < msgLength - CRC_SIZE; i++) {
      validityBytes[i - 1] = payloadBuffer[i];
    }
    Serial1.print("validity from host: ");
    char charBuffer[2];
    for (uint8_t i = 0; i < msgLength - CRC_SIZE - 1; i++) {
      Serial1.print(validityBytes[i], HEX);
      sprintf(charBuffer, "%d", validityBytes[i]);
      validityString += String(validityBytes[i]);//String(deviceIDBytes[i], 16);
    }

    Serial1.println();
    Serial1.print("set Cert validity to ");
    Serial1.println(validityString);
    Cert.setExpireYears(validityString.toInt());

    char response[] = {char(RESPONSE::RESPONSE_ACK)};
    sendData(MESSAGE_TYPE::RESPONSE, response, 1);

  }

  if (cmdCode == COMMAND::SET_CERT_SERIAL) {
    // extract payload from [1] to [payloadLength]
    // this will be the certificate serial number included in the device certificate
    Serial1.println("set CERT Serial");
    byte certSerialBytes[msgLength - CRC_SIZE - 1];

    for (uint8_t i = 1; i < msgLength - CRC_SIZE; i++) {
      certSerialBytes[i - 1] = payloadBuffer[i];
    }

    Serial1.print("Serial Number from host: ");
    for (uint8_t i = 0; i < msgLength - CRC_SIZE - 1; i++) {
      Serial1.print(certSerialBytes[i], HEX);
    }
    Serial1.println();

    if(!Cert.setSerialNumber(certSerialBytes, sizeof(certSerialBytes))) {
      Serial1.println("set CERT Error");
      char response[] = {char(RESPONSE::RESPONSE_ERROR)};
      sendData(MESSAGE_TYPE::RESPONSE, response, 1);
      return;
    }

    char response[] = {char(RESPONSE::RESPONSE_ACK)};
    sendData(MESSAGE_TYPE::RESPONSE, response, 1);
  }
  if (cmdCode == COMMAND::SET_AUTH_KEY) {
    // extract payload from [1] to [payloadLength]
    // this will be the authority key id included in the device certificate
    Serial1.println("set Auth Key Id");
    byte authKeyBytes[msgLength - CRC_SIZE - 1];

    for (uint8_t i = 1; i < msgLength - CRC_SIZE; i++) {
      authKeyBytes[i - 1] = payloadBuffer[i];
    }

    Serial1.print("Authority Key Id from host: ");
    for (uint8_t i = 0; i < msgLength - CRC_SIZE - 1; i++) {
      Serial1.print(authKeyBytes[i], HEX);
    }
    Serial1.println();

    if(!Cert.setAuthorityKeyId(authKeyBytes, sizeof(authKeyBytes))) {
      Serial1.println("set AuthorityKeyId Error");
      char response[] = {char(RESPONSE::RESPONSE_ERROR)};
      sendData(MESSAGE_TYPE::RESPONSE, response, 1);
      return;
    }

    char response[] = {char(RESPONSE::RESPONSE_ACK)};
    sendData(MESSAGE_TYPE::RESPONSE, response, 1);
  }
  if (cmdCode == COMMAND::SET_SIGNATURE) {
    // extract payload from [1] to [payloadLength]
    // this will be the signature included in the device certificate
    Serial1.println("set Signature ");
    byte signatureBytes[msgLength - CRC_SIZE - 1];

    for (uint8_t i = 1; i < msgLength - CRC_SIZE; i++) {
      signatureBytes[i - 1] = payloadBuffer[i];
    }
    Serial1.println();

    Serial1.print("Signature from host: ");
    for (uint8_t i = 0; i < msgLength - CRC_SIZE - 1; i++) {
      Serial1.print(signatureBytes[i], HEX);
    }
    Serial1.println();

    if(!Cert.setSignature(signatureBytes, sizeof(signatureBytes))) {
      Serial1.println("set signature Error");
      char response[] = {char(RESPONSE::RESPONSE_ERROR)};
      sendData(MESSAGE_TYPE::RESPONSE, response, 1);
      return;
    }

    char response[] = {char(RESPONSE::RESPONSE_ACK)};
    sendData(MESSAGE_TYPE::RESPONSE, response, 1);
  }
  if (cmdCode == COMMAND::END_STORAGE) {
    Serial1.println("end storage");
    if (!Crypto.buildCert(Cert, CryptoSlot::Key)) {
      Serial1.println("Error creating cert!");
      char response[] = {char(RESPONSE::RESPONSE_ERROR)};
      sendData(MESSAGE_TYPE::RESPONSE, response, 1);
      return;
    }

    if (!Crypto.writeCert(Cert, CryptoSlot::CompressedCertificate)) {
      Serial1.println("Error storing Crypto cert!");
      char response[] = {char(RESPONSE::RESPONSE_ERROR)};
      sendData(MESSAGE_TYPE::RESPONSE, response, 1);
      return;
    }

    Serial1.println("Certificate PEM file");
    Serial1.println(Cert.getCertPEM());

    char response[] = {char(RESPONSE::RESPONSE_ACK)};
    sendData(MESSAGE_TYPE::RESPONSE, response, 1);
  }


  if (cmdCode == COMMAND::RECONSTRUCT_CERT) {

    if (!Cert.begin()) {
      Serial1.println("Error starting Crypto cert reconstruction!");
      char response[] = {char(RESPONSE::RESPONSE_ERROR)};
      sendData(MESSAGE_TYPE::RESPONSE, response, 1);
      return;
    }

    if (!Crypto.readCert(Cert, CryptoSlot::CompressedCertificate)) {
      Serial1.println("Error reconstructing Crypto cert!");
      char response[] = {char(RESPONSE::RESPONSE_ERROR)};
      sendData(MESSAGE_TYPE::RESPONSE, response, 1);
      return;
    }

    Serial1.println("Compressed cert = ");

    const byte *certData = Cert.bytes();
    int certLength = Cert.length();

    for (int i = 0; i < certLength; i++) {
      byte b = certData[i];

      if (b < 16) {
        Serial1.print('0');
      }
      Serial1.print(b, HEX);

    }
    Serial1.println();
    char response[] = {char(RESPONSE::RESPONSE_ACK)};
    sendData(MESSAGE_TYPE::RESPONSE, response, 1);
  }

}

void processRawData(bool checkSumOK) {
  Serial1.println(">> processing raw data");

  if (checkSumOK) {
    uint8_t resp[] = {0x55, 0xaa, (uint8_t)MESSAGE_TYPE::RESPONSE, 0x01, (uint8_t)RESPONSE::RESPONSE_ACK, 0xaa, 0x55};
    for (uint8_t i = 0; i < sizeof(resp); i++) {
      Serial1.print(resp[i]);
      Serial1.print(" ");
    }
    Serial.write(resp, sizeof(resp));
  } else {
    uint8_t resp[] = {0x55, 0xaa, (uint8_t)MESSAGE_TYPE::RESPONSE, 0x01, (uint8_t)RESPONSE::RESPONSE_NACK, 0xaa, 0x55};
    for (uint8_t i = 0; i < sizeof(resp); i++) {
      Serial1.print(resp[i]);
      Serial1.print(" ");
    }
    Serial.write(resp, sizeof(resp));
  }
}

void sendData(MESSAGE_TYPE _msgType, const char* _msgPayload, uint16_t _payloadSize) {
  Serial1.print("payload size: ");
  Serial1.print(_payloadSize);
  Serial1.print(" [");
  Serial1.print(_payloadSize, HEX);
  Serial1.println("]");
  Serial1.println(_msgPayload);

  Serial.write(msgStart, sizeof(msgStart));
  Serial.write((uint8_t)_msgType);
  Serial.write((_payloadSize + CRC_SIZE) >> 8) ;
  Serial.write((_payloadSize + CRC_SIZE) & 0xff);
  Serial.write(_msgPayload, _payloadSize);
  uint16_t payloadCRC = uCRC16Lib::calculate((char *)_msgPayload, _payloadSize);
  Serial1.print("payload CRC out: ");
  Serial1.println(payloadCRC, HEX);
  Serial.write((uint8_t)(payloadCRC >> 8));
  Serial.write((uint8_t)(payloadCRC & 0xff));
  Serial.write(msgEnd, sizeof(msgEnd));

}

void sendResponse() {

}
void changeState(MACHINE_STATE _newState) {
  Serial1.print("changing state to ");
  Serial1.println((uint8_t)_newState);
  if (_newState == machineState)
    return;
  if (_newState == MACHINE_STATE::RECEIVING_PAYLOAD) {
    msgByteIndex = 0;
  }
  if (_newState == MACHINE_STATE::IDLE) {

  }
  machineState = _newState;
}

uint8_t cryptoInit() {
  unsigned long ecctimeout = 1000;
  unsigned long beginOfTime = millis();
  bool eccOK = 0;
  while (!(eccOK = Crypto.begin()) || (millis() - beginOfTime < ecctimeout)) {
  }
  return eccOK;
}

PROVISIONING_ERROR cryptoLock() {
  if (!Crypto.locked()) {

    if (!Crypto.writeConfiguration(DEFAULT_ECCX08_TLS_CONFIG)) {
      return PROVISIONING_ERROR::WRITE_CONFIG_FAIL;
    }

    if (!Crypto.lock()) {
      return PROVISIONING_ERROR::LOCK_FAIL;
    }
    return PROVISIONING_ERROR::LOCK_SUCCESS;
  }
  return PROVISIONING_ERROR::LOCK_SUCCESS;
}

PROVISIONING_ERROR generateCSR() {
  if (!Crypto.locked()) {
    Serial1.println("Chip is not locked");
    return PROVISIONING_ERROR::LOCK_FAIL;
  }
  Serial1.println("CSR generation in progress");

  if (!Cert.begin()) {
    Serial1.println("Error initializing CSR data!");
    return PROVISIONING_ERROR::CSR_GEN_FAIL;
  }

  Cert.setSubjectCommonName(deviceIDstring);

  if (!Crypto.buildCSR(Cert, CryptoSlot::Key, true)) {
    Serial1.println("Error generating CSR!");
    return PROVISIONING_ERROR::CSR_GEN_FAIL;
  }

  csr = Cert.getCSRPEM();
  if (!csr) {
    Serial1.println("Error reading CSR PEM data!");
    return PROVISIONING_ERROR::CSR_GEN_FAIL;
  }
  Serial1.println(csr.length());
  Serial1.println(csr);

  return PROVISIONING_ERROR::CSR_GEN_SUCCESS;
}
