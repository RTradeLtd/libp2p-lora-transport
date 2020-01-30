package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jacobsa/go-serial/serial"
)

// Bridge enables a LibP2P node to communicate over LoRa
type Bridge struct {
	serial     io.ReadWriteCloser
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
	config := serial.OpenOptions{
		PortName:        "/dev/ttyACM0",
		BaudRate:        115200,
		DataBits:        8,
		ParityMode:      0,
		StopBits:        1,
		MinimumReadSize: 4,
		//	ReadTimeout: time.Second * 10,
	}
	sh, err := serial.Open(config)
	if err != nil {
		log.Fatal(err)
	}
	defer sh.Close()
	doneChan := make(chan bool, 1)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go handleExit(ctx, cancel, wg, doneChan)
	_, err = sh.Write([]byte("1"))
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 5)
	for {
		select {
		case <-ctx.Done():
			goto FIN
		default:
			data := make([]byte, 255)
			_, err := sh.Read(data)
			if err != nil && err != io.EOF {
				fmt.Println("failed to read data: ", err.Error())
				goto FIN
			}
			msg, err := bufio.NewReader(sh).ReadString('\n')
			if err != nil {
				fmt.Println("failed to read data: ", err.Error())
				goto FIN
			}
			fmt.Println(msg)
		}
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
