package bridge

import (
	"context"
	"errors"
	"sync"
	"testing"

	"go.uber.org/zap/zaptest"
)

func Test_SerialDumper(t *testing.T) {
	fserial := NewFakeSerial()
	logger := zaptest.NewLogger(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	bridge, err := NewBridge(ctx, &sync.WaitGroup{}, logger, fserial, Opts{})
	if err != nil {
		t.Fatal(err)
	}
	s, err := fserial.Write([]byte("^hello^"))
	if err != nil {
		t.Fatal(err)
	}
	if s != len("^hello^") {
		t.Fatal("err")
	}
	// cause a trigger of the "write loop"
	bridge.writeChan <- []byte("^hello^")
	data := <-bridge.readChan
	if string(data) != "^hello^" {
		t.Fatal("bad test data")
	}
}

var _ Serial = (*FakeSerial)(nil)

// FakeSerial implements fake serial
type FakeSerial struct {
	mx          sync.RWMutex
	errNextCall bool
	nextErr     error
	nextRead    []byte
}

func NewFakeSerial() *FakeSerial {
	return &FakeSerial{}
}

func (fs *FakeSerial) ToggleError(err error) {
	fs.errNextCall = !fs.errNextCall
}

func (fs *FakeSerial) Write(data []byte) (int, error) {
	if fs.errNextCall {
		return 0, errors.New("error")
	}
	fs.nextRead = data
	return len(data), nil
}

func (fs *FakeSerial) Available() (int, error) {
	if fs.errNextCall {
		return 0, errors.New("error")
	}
	return len(fs.nextRead), nil
}

func (fs *FakeSerial) Read(data []byte) (int, error) {
	if fs.errNextCall {
		return 0, errors.New("error")
	}
	copy(data, fs.nextRead)
	fs.nextRead = nil
	return len(data), nil
}

func (fs *FakeSerial) Flush() error {
	if fs.errNextCall {
		return errors.New("error")
	}
	return nil
}

func (fs *FakeSerial) Close() error {
	if fs.errNextCall {
		return errors.New("error")
	}
	return nil
}
