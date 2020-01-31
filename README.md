# libp2p-lora-transport

`libp2p-lora-transport` enables LibP2P nodes to communicate over LoRa. You can either use it as a "protocol" where a LibP2P nodes with an attached LoRa bridge can allow authorized peers to read/write data from/to the LoRa bridge. For example, this could be used to allow a LibP2P nodes to report sensor data to a LoRaWAN gateway. Another possibility would be to allow multiple different LibP2P nodes to relay data through a LibP2P node with an attached LoRa bridge.

# Hardware

The following hardware has been tested:

* Arduino Mega 2560 Rev3 + [Dragino LoRa Shield Rev 1.4](http://wiki.dragino.com/index.php?title=Lora_Shield)


# Architecture

Using an Arduino Mega + Dragino LoRa GPS shield, a sketch called "lora bridge" is deployed to the arduino. This sketch is responsible for two things:

* Take data coming in on the serial interface, and push it out the LoRa interface
* Take data coming in on the LoRa interface, and push it out the serial interface

A LibP2P host with a direct connection to the arduino serial interface registers a bridge handler that connects to the arduino. This bridge handler creates two channels, one for writing data into the serial interface, one for reading data out of the serial interface. A goroutine is then launched, which will pull data off the write channel, and pipe it into the serial interface. If no data is available for writing, we then see if any data can be read off the serial interface. If we can, we read the data, and send it through the read channel. If no one is waiting to receive from the channel, the data is simply discarded.

The bridge will ensure that all messages coming off the serial interface are properly formatted (wrapped in carrats `^`), if the messages aren't they are also discarded.

## Serial Communications

The serial interface on the arduino is used to allow our LibP2P nodes access to the LoRa module. Anytime data is sent from the arduino to the LibP2P node, messages are wrapped in `^`. For example, should we wish to send a message to another LoRa node thats says `hello` we should send `^hello^`. Controlling the LoRa bridge is done via single letter "control characters". The current control characters are:

* `1` - Toggle debug mode
  * Debug mode switches processing so that anytime a LoRa packet is received, debug information is sent thorugh the serial interface containing the RSSI, SNR, and Error Frequency of the LoRa packet.

# License

All non firmata code is licensed under AGPLv3