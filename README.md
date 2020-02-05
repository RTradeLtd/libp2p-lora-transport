# libp2p-lora-transport

`libp2p-lora-transport` enables LibP2P nodes to communicate over LoRa. You can either use it as a "protocol" where a LibP2P nodes with an attached LoRa bridge can allow authorized peers to read/write data from/to the LoRa bridge. For example, this could be used to allow a LibP2P nodes to report sensor data to a LoRaWAN gateway. Another possibility would be to allow multiple different LibP2P nodes to relay data through a LibP2P node with an attached LoRa bridge.

# Install

You will need a valid C installation, Go 1.13+, and the WiringPi libraries installed.

# Hardware

The following hardware has been tested:

* Arduino Mega 2560 Rev3 and [Dragino LoRa Shield Rev 1.4](http://wiki.dragino.com/index.php?title=Lora_Shield)
* Raspberry Pi 3B+ and [Dragino LoRa GPS HAT](http://wiki.dragino.com/index.php?title=Lora/GPS_HAT)

# Contents

* `arduino`
  * Contains arduino related sketch files, and includes the LoRa shield bridge
  * Also includes a copy of the firmata sketch, and a heltec esp32 LoRa board compatible sketch
* `bridge`
  * LibP2P golang code providing a transport, and protocol capable of talking over LoRa by way of a directly connected LoRa bridge
  * Alternatively you will be able to leverage this code, along with the Dragino LoRa GPS HAT to enable LibP2P IPFS nodes that can facilitate swarm communication over LoRa
  * It is unlikely you will be able to use the LoRa shield bridge, as this requires a serial port connection which is slower
  * Additionally the arduino is pretty resource constrained and we will be able to achieve higher throughput with the raspberry pi
* `include`
  * various C header files
* `src/dragino`
  * Contains a modified version of the sample code provided with the Dragino LoRa GPS HAT intended to be used as a library, and easier to maintain
  * Additionally there is a CGO file allowing golang programs to use the dragino HAT.
  * A cgo library for using the dragino lora gps hat from go programs

# Architecture

Using an Arduino Mega + Dragino LoRa GPS shield, a sketch called "lora bridge" is deployed to the arduino. This sketch is responsible for two things:

* Take data coming in on the serial interface, and push it out the LoRa interface
* Take data coming in on the LoRa interface, and push it out the serial interface

A LibP2P host with a direct connection to the arduino serial interface registers a bridge handler that connects to the arduino. This bridge handler creates two channels, one for writing data into the serial interface, one for reading data out of the serial interface. A goroutine is then launched, which will pull data off the write channel, and pipe it into the serial interface. If no data is available for writing, we then see if any data can be read off the serial interface. If we can, we read the data, and send it through the read channel. If no one is waiting to receive from the channel, the data is simply discarded. 

The bridge will ensure that all messages coming off the serial interface are properly formatted (wrapped in carrats `^`), if the messages aren't they are also discarded.

There are two modes of operation:

* Protocol Mode (libp2p protocol that can be accessed through libp2p streams)
* Transport (used as an actual swarm transport, TODO).

## Security

There is absolutely no security provided in this implementation. Data is handled as is, and if that data is in cleartext, then data will be transmitted over the LoRa radio in cleartext for anyone whose listening to snoop on. That means if you want data to be private going through this bridge, you must encrypt it. If using this bridge as a transport (non protocol mode) then it's recommended to use a private libp2p swarm as that provides a reasonably good base layer of security, without having to manually encrypt the data going through the bridge. If using the bridge in protocol mode this means you will need to encrypt the messages manually.


In protocol mode any authorized peer can read/write data through the bridge, so make sure that you only allow particular peers access.

## Serial Communications

The serial interface on the arduino is used to allow our LibP2P nodes access to the LoRa module. Anytime data is sent from the arduino to the LibP2P node, messages are wrapped in `^`. For example, should we wish to send a message to another LoRa node thats says `hello` we should send `^hello^`. Controlling the LoRa bridge is done via single letter "control characters". The current control characters are:

* `1` - Toggle debug mode
  * Debug mode switches processing so that anytime a LoRa packet is received, debug information is sent thorugh the serial interface containing the RSSI, SNR, and Error Frequency of the LoRa packet.

# License

All non firmata code is licensed under AGPLv3

# Notes:

* https://dave.cheney.net/tag/cgo
* https://golang.org/cmd/cgo/