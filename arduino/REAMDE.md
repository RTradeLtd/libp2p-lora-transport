# arudino

Contains arduino sketches. Requires master branch of `https://github.com/sandeepmistry/arduino-LoRa`

* `custom_firmata.ino` is a custom version of firmata that starts up the LoRa radio and allows controlling via gobot (TODO)
* `firmata.ino` is a copy of the firmata sketch
* `lora_bridge.ino` is a "debuggable" version of the LoRa bridge
* `lora_bridge_fast.ino` is an optimize version of the LoRa bridge, focused on getting data in/out as fast as possible

# bridge

The bridge currently has some issues reading data from the arduino serial interface, see https://stackoverflow.com/questions/50088669/golang-reading-from-serial for more information.
Trying with a new library and still getting weirdness:

```
rssi: -70	snr: 9.75	errFreq: -3279
rssi: -76	snr: 10.00	errFreq: -3279

rssi: -68	snr: 10.7
5	errFreq: -3279

rssi: -68	snr: 9.75	errFreq: -3279

rssi: -70	snr: 10.00	errFreq: -3263

rssi: -73	snr: 
9.25	errFreq: -3263
```