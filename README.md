# libp2p-lora-transport

This is an experimental libp2p transport for communicating over LoRa, and possibly LoRaWAN. It is meant to be used with the Dragino LoRa GPS shields. And is capable of receiving data on LoRa and sending it through the serial interface, or reading from the serial interface and sending it thorugh LoRa. Essentially the arduino's act as a bridge between LoRa and LibP2P.

# Architecture

An Arduino Mega + Dragino LoRa GPS Shield, or ES32 LoRa chips such as those from heltec run the "lora bridge" sketch contained in the `arduino` folder. This sketch serves two purposes:

* Take data coming in on the serial interface, and push it out the LoRa interface
* Take data coming in on the LoRa interface, and push it out the serial interface

Using these two functions we can then implement a LibP2P transport, satisfying the "Write" and "Read" interfaces, where "Read" means to read data coming out the serial interface, and "Write" means to write data out the LoRa interface.