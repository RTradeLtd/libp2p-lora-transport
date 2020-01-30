#include <LoRa.h>
#include <util/atomic.h> // this library includes the ATOMIC_BLOCK macro.
/* LibP2P LoRa Transport Arduino Bridge
   Enables a LibP2P node to send messages over LoRa and possibly LoRaWAN.
   Data coming in through the serial interface leaves through the LoRa radio
   Data coming in through the LoRa radio exits the serial interface
   This is the "fast" version and operates on bytes instead of strings
*/

#define BAND 915E6
#define SYNC 0x99
#define SF 7
#define TXPOWER 20
#define SIGBW 250E3
#define DATASENT 0xd34db33f

// means this can be accessed by other parts of the program
volatile bool debug; 

/*
  API documentation: https://github.com/sandeepmistry/arduino-LoRa/blob/master/API.md
  Issues with timing: https://github.com/sandeepmistry/arduino-LoRa/issues/321
  Other example:
    https://github.com/IoTThinks/EasyLoRaGateway_v2.1/blob/master/EasyLoRaGateway/09_lora.ino
    Useful doc https://www.arduino.cc/en/Tutorial/SerialCallResponse
*/

void setup() {
  Serial.begin(2500000);
  while (!Serial);
  LoRa.begin(BAND);
  LoRa.setSyncWord(SYNC);
  LoRa.setSpreadingFactor(SF);
  LoRa.setTxPower(TXPOWER);
  // register callbacks
  LoRa.onReceive(onReceive);
  LoRa.onTxDone(onTxDone);
  LoRa.enableCrc();
  // LoRa.setSignalBandwidth(SIGBW);
  LoRa.receive(); // set receive mode, allows using callback
  Serial.println("ready");
  Serial.flush();
}

// callback function whenever we receive a LoRa packet
void onReceive(int packetSize) {
  if (packetSize) {
    if (debug) {
      Serial.println("rssi: " + String(LoRa.packetRssi()) + " snr: " + String(LoRa.packetSnr()) + " errFreq: " + String(LoRa.packetFrequencyError()) + "\t");
      Serial.flush();
      return;          
    }
    char buffer[255];
    int num = LoRa.readBytes(buffer, 255);
    Serial.write(buffer, num);
    Serial.println();
    Serial.flush();
  }
}

// callback function whenever we finished sending a LoRa packet
void onTxDone() {
  Serial.println(DATASENT);
//  Serial.flush();
  LoRa.receive(); // put us back in receive mode to allow usage of call back
}

void loop() {
  // see if we have any data on the serial interface
  // if we do send it down hte LoRa radio
  if (Serial.available()) {
    char buffer[255]; // 255 byte buffer
    int num = Serial.readBytesUntil('\n', buffer, 255);
    switch (num) {
      case 1:
        if (buffer[0] == '1') { // debug mode, otherwise fallthrough
          ATOMIC_BLOCK(ATOMIC_RESTORESTATE) {
            // code with interrupts blocked (consecutive atomic operations will not get interrupted)
            // https://www.arduino.cc/reference/en/language/variables/variable-scope--qualifiers/volatile/
            debug = !debug; 
          }
          break;
        }
      default:
        if (num >= 1) {
          if (LoRa.beginPacket()) { // beginPacket returns 1 if ready to proceed, so wait until radio is ready
            LoRa.write(buffer, num);
            LoRa.endPacket(true); // async mode
          }
        }
    }
  }
}