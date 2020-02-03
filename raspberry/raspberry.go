package raspberry

/*
#cgo CFLAGS: -c -Wall
#cgo LDFLAGS: -lwiringPi -lpthread
#include <stdio.h>
#include <stdbool.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <string.h>
#include <sys/time.h>
#include <signal.h>
#include <stdlib.h>
#include <sys/ioctl.h>
#include <wiringPi.h>
#include <wiringPiSPI.h>
#include <raspberry.c>
*/
import "C"

// Bridge is a raspberrypi dragino lora bridge
type Bridge struct {
	isSender bool
}

// NewBridge initializes our Dragino LoRa GPS HAT
func NewBridge(isSender bool) (*Bridge, error) {
	_, err := C.Setup(C.bool(isSender))
	if err != nil {
		return nil, err
	}
	return &Bridge{},nil
}
