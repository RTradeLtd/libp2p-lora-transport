package bridge

import (
	"context"
	"sync"
	"testing"

	"github.com/pkg/term"
	"github.com/pkg/term/termios"
	"go.uber.org/zap/zaptest"
)

func devTerm(t *testing.T) *term.Term {
	ptm, pts, err := termios.Pty()
	if err != nil {
		t.Fatal(err)
	}
	ptm.Write([]byte("^hello^"))
	pts.Write([]byte("^hello^"))
	trm, err := term.Open(pts.Name())
	if err != nil {
		t.Fatal(err)
	}
	//pts.Close()
	return trm
}

func Test_SerialDumper(t *testing.T) {
	trm := devTerm(t)
	defer trm.Close()
	logger := zaptest.NewLogger(t)
	bridge, err := NewBridge(context.Background(), &sync.WaitGroup{}, logger, trm, Opts{})
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		_, err := bridge.serial.Write([]byte("^hello^"))
		if err != nil {
			t.Error(err)
		}
	}()
	data := <-bridge.readChan
	if string(data) != "hello" {
		t.Fatal("bad test data")
	}
}
