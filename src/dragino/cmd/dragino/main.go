package main

import (
	"log"

	raspberry "github.com/RTradeLtd/libp2p-lora-transport/src/dragino"
)

func main() {
	_, err := raspberry.NewBridge(true)
	if err != nil {
		log.Fatal(err)
	}
	raspberry.WriteBridge([]byte("hello"))
}
