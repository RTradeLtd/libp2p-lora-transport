# libp2p-lora-transport

`libp2p-lora-transport` enables LibP2P nodes to communicate over LoRa. You can either use it as a "protocol" where a LibP2P nodes with an attached LoRa bridge can allow authorized peers to read/write data from/to the LoRa bridge. For example, this could be used to allow a LibP2P nodes to report sensor data to a LoRaWAN gateway. Another possibility would be to allow multiple different LibP2P nodes to relay data through a LibP2P node with an attached LoRa bridge.

# Hardware

The following hardware has been tested:

* Arduino Mega 2560 Rev3 + [Dragino LoRa Shield Rev 1.4](http://wiki.dragino.com/index.php?title=Lora_Shield)


# Architecture

An Arduino Mega + Dragino LoRa GPS Shield, or ES32 LoRa chips such as those from heltec run the "lora bridge" sketch contained in the `arduino` folder. This sketch serves two purposes:

* Take data coming in on the serial interface, and push it out the LoRa interface
* Take data coming in on the LoRa interface, and push it out the serial interface

Using these two functions we can then implement a LibP2P transport, satisfying the "Write" and "Read" interfaces, where "Read" means to read data coming out the serial interface, and "Write" means to write data out the LoRa interface.

## Possible Variations:

* Instead of a transport, have it be a "protocol", where supporting hosts can provide access to a type of protocol, that allows reading/writing from the LoRa interface