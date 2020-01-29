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

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	config := &serial.Config{
		Name: "/dev/ttyACM0",
		Baud: 9600,
	}
	sh, err := serial.OpenPort(config)
	if err != nil {
		log.Fatal(err)
	}
	doneChan := make(chan bool, 1)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go handleExit(ctx, cancel, wg, doneChan)
	for {
		select {
		case <-ctx.Done():
			goto FIN
		default:
			_, err := sh.Write([]byte("hello world"))
			if err != nil && err != io.EOF {
				fmt.Println("failed to write data: ", err.Error())
				goto FIN
			}
			data := make([]byte, 1024)
			_, err = sh.Read(data)
			if err != nil && err != io.EOF {
				fmt.Println("failed to read data: ", err.Error())
				goto FIN
			}
			if data != nil {
				fmt.Println(string(data))
			}
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
