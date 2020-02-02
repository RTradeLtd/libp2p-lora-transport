package main

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
import (
	"log"
)

func main() {
	_, err := C.Setup(true)
	if err != nil {
		log.Fatal(err)
	}
}
