#include <SPI.h>
#include <LoRa.h>

/* LibP2P LoRa Transport Arduino Bridge
 * Enables a LibP2P node to send messages over LoRa and possibly LoRaWAN.
 * Data coming in through the serial interface leaves through the LoRa radio
 * Data coming in through the LoRa radio exits the serial interface
 */

#define BAND 915E6
#define SYNC 0x99
#define SF 7
#define TXPOWER 20
#define SIGBW 250E3
#define DATASENT 0xd34db33f
/*
API documentation: https://github.com/sandeepmistry/arduino-LoRa/blob/master/API.md
*/

void setup() {
    Serial.begin(115200);
    while (!Serial);
    LoRa.begin(BAND);
    LoRa.setSyncWord(SYNC);
    LoRa.setSpreadingFactor(SF);
    LoRa.setTxPower(TXPOWER);
    //LoRa.enableCrc();
  //  LoRa.setSignalBandwidth(SIGBW);
    Serial.println("ready");
    Serial.flush();
}

// allows debugging LoRa packet note that this is a blocking call
// and any other usage of the bridge will be blocked until then.
void handleDebug() {
  while (true) {
    if (LoRa.parsePacket()) {
        String message = "received packet: rssi " + String(LoRa.packetRssi()) + " snr " + String(LoRa.packetSnr()) + " freq error " + String(LoRa.packetFrequencyError());
        Serial.println(message);
        Serial.flush();
        return;
    }
  }
}

void loop() {
  // see if we have any data on the LoRa radio
  // if we do, send it down the serial interface
  if (LoRa.parsePacket()) {
    Serial.println(LoRa.readString());
    Serial.flush();
  }
  // see if we have any data on the serial interface
  // if we do send it down hte LoRa radio
  if (Serial.available()) {
    if (LoRa.beginPacket()) { // beginPacket returns 1 if ready to proceed, so wait until radio is ready
        String msg = Serial.readStringUntil('\n');
        if (msg == "debug") {
          handleDebug();
        }
        LoRa.print(msg);
        LoRa.endPacket(true); // async mode
        Serial.println(msg);
        Serial.println(DATASENT);
        Serial.flush();
    } 
  }
}