package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/firmata"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	firmataAdaptor := firmata.NewAdaptor("/dev/ttyACM0")

	robot := gobot.NewRobot("bot",
		[]gobot.Connection{firmataAdaptor},
	)
	doneChan := make(chan bool, 1)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go handleExit(ctx, cancel, wg, doneChan)
	robot.Start()
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-doneChan
		robot.Stop()
	}()
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
