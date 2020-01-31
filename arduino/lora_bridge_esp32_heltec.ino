#include "heltec.h"

#define BAND    915E6  //you can set band here directly,e.g. 868E6,915E6
#define SYNC 0x99
#define SF 7
#define TXPOWER 20
#define SIGBW 250E3

volatile bool debug;

void setup() {
  //WIFI Kit series V1 not support Vext control
  Heltec.begin(false /*DisplayEnable Enable*/, true /*Heltec.Heltec.Heltec.LoRa Disable*/, true /*Serial Enable*/, true /*PABOOST Enable*/, BAND /*long BAND*/);
  Serial.begin(115200);
  while (!Serial);
  LoRa.setSyncWord(SYNC);
  LoRa.setFrequency(BAND);
  LoRa.setSpreadingFactor(SF);
  LoRa.enableCrc();
  LoRa.setTxPower(TXPOWER, RF_PACONFIG_PASELECT_PABOOST);
  LoRa.onReceive(onReceive);
  LoRa.receive();
  Serial.println("ready");
  Serial.flush();
}


// callback function whenever we receive a LoRa packet
void onReceive(int packetSize) {
  if (packetSize) {
    if (debug) {
      Serial.print("^" + String(LoRa.packetRssi()) + "," + String(LoRa.packetSnr()) + "^");
      Serial.flush();
      return;
    }
    char buffer[255];
    Serial.print("^");
    int num = LoRa.readBytes(buffer, 255);
    Serial.print(buffer);
    Serial.print("^");
    Serial.flush();
  }
}


// callback function whenever we finished sending a LoRa packet
void onTxDone() {
  LoRa.receive(); // put us back in receive mode to allow usage of call back
}

// https://www.arduino.cc/reference/en/language/functions/communication/serial/serialevent/
void serialEvent() {
  // see if we have any data on the serial interface
  // if we do send it down hte LoRa radio
  if (Serial.available()) {
    char buffer[255]; // 255 byte buffer
    int num = Serial.readBytesUntil('\n', buffer, 255);
    switch (num) {
      case 1:
        if (buffer[0] == '1') { // debug mode, otherwise fallthrough
          debug = !debug;
          break;
        }
      default:
        if (num >= 1) {
          if (LoRa.beginPacket()) { // beginPacket returns 1 if ready to proceed, so wait until radio is ready
            LoRa.print(buffer);
            LoRa.endPacket(); // async mode
            onTxDone();
          }
        }
    }
  }
}

void loop() {}