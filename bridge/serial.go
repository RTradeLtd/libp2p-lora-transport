package bridge

import (
	"errors"
	"sync"

	"github.com/pkg/term"
)

var _ Serial = (*term.Term)(nil)
var _ Serial = (*FakeSerial)(nil)

// Serial implements serial capable interfaces
type Serial interface {
	Write([]byte) (int, error)
	Available() (int, error)
	Read([]byte) (int, error)
	Flush() error
	Close() error
}

type FakeSerial struct {
	mx          sync.RWMutex
	errNextCall bool
	nextRead    []byte
}

func NewFakeSerial() *FakeSerial {
	return &FakeSerial{}
}

func (fs *FakeSerial) ToggleError() {
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
	for i := range data {
		if i <= len(fs.nextRead) {
			data[i] = fs.nextRead[i]
		} else {
			break
		}
	}
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
