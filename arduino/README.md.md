# arudino

Contains arduino sketches. Requires master branch of `https://github.com/sandeepmistry/arduino-LoRa`

* For the firmata sketch file see [here](https://gist.github.com/bonedaddy/3c1bf4ff93e4b72372ac2806a24b203d)
* `lora_bridge_fast.ino` LoRa bridge for LibP2P nodes
* `lora_bridge_esp32_heltec.ino` is the LoRa bridge for ESP32 Heltec boards

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

I think the above issue might be related to `baud` issues?? trying with a `2500000` baud vs a `115200` results in the error happening a lot less