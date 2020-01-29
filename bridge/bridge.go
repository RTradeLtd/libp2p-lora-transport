package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/tarm/serial"
)

// Bridge enables a LibP2P node to communicate over LoRa
type Bridge struct {
	serial     *serial.Port
	ctx        context.Context
	cancelFunc context.CancelFunc
	mx         *sync.RWMutex
}

// Read processes data coming in from the serial interface
func (b *Bridge) Read() {
	b.mx.RLock()
	data := make([]byte, 1024)
	read, err := b.serial.Read(data)
	if err != nil && err != io.EOF {
		panic(err)
	}
	if read > 0 {
		fmt.Println(string(data[:read]))
	}
	b.mx.RUnlock()
}

// Write sends data through the serial interface
func (b *Bridge) Write(data []byte) (int, error) {
	b.mx.Lock()
	n, err := b.serial.Write(data)
	b.mx.Unlock()
	return n, err
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	config := &serial.Config{
		Name: "/dev/ttyACM0",
		Baud: 115200,
	}
	sh, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal(err)
	}
	doneChan := make(chan bool, 1)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go handleExit(ctx, cancel, wg, doneChan)
	var (
		dataBuffer []byte
		delim      int
	)
	for {
	START:
		select {
		case <-ctx.Done():
			goto FIN
		default:
			data := make([]byte, 1024)
			read, err := sh.Read(data)
			if err != nil && err != io.EOF {
				fmt.Println("failed to read data: ", err.Error())
				goto FIN
			}
			if read > 0 {
				dataBuffer = append(dataBuffer, data[:read]...)
				count := 0
				for _, c := range dataBuffer {
					if c == '\n' {
						delim = count
						goto PRINT
					}
					count++
				}
			}
		}
	PRINT:
		var trimmed []byte
		for _, c := range dataBuffer[:delim] {
			if c != ' ' {
				trimmed = append(trimmed, c)
			}
		}
		fmt.Println(string(trimmed))
		dataBuffer = []byte{}
		goto START
	}
FIN:
	cancel()
	wg.Wait()
}

func handleExit(ctx context.Context, cancelFunc context.CancelFunc, wg *sync.WaitGroup, doneChan chan bool) {
	defer wg.Done()
	// make a channel to catch os signals on
	quitCh := make(chan os.Signal, 1)
	// register the types of os signals to trap
	signal.Notify(quitCh, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	// wait until we receive an exit signal
	<-quitCh
	// cancel the context which will trigger shutdown of service components
	cancelFunc()
	// notify that we are finished handling all exit procedures
	doneChan <- true
}
