package bridge

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	testutils "github.com/RTradeLtd/go-libp2p-testutils"
	"github.com/libp2p/go-libp2p-core/crypto"
	host "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
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

func Test_StreamHandler(t *testing.T) {
	wg := &sync.WaitGroup{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	secret := testutils.NewSecret(t)
	h1, _ := newHost(ctx, t, "/ip4/127.0.0.1/tcp/4005", secret)
	defer h1.Close()
	h2, _ := newHost(ctx, t, "/ip4/127.0.0.1/tcp/4006", secret)
	defer h2.Close()
	h1.ConnManager().Protect(h2.ID(), "test")
	h2.ConnManager().Protect(h1.ID(), "test")
	for _, addr := range h1.Addrs() {
		fmtAddr := fmt.Sprintf("%s/p2p/%s", addr.String(), h1.ID())
		ma, err := multiaddr.NewMultiaddr(fmtAddr)
		if err != nil {
			t.Fatal(err)
		}
		h2.Peerstore().AddAddr(h1.ID(), ma, time.Hour)
		peerInfo, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			t.Fatal(err)
		}
		h2.Connect(ctx, *peerInfo)
	}
	for _, addr := range h2.Addrs() {
		fmtAddr := fmt.Sprintf("%s/p2p/%s", addr.String(), h2.ID())
		ma, err := multiaddr.NewMultiaddr(fmtAddr)
		if err != nil {
			t.Fatal(err)
		}
		h1.Peerstore().AddAddr(h2.ID(), ma, time.Hour)
		peerInfo, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			t.Fatal(err)
		}
		h1.Connect(ctx, *peerInfo)
	}
	h1.Connect(ctx, h2.Peerstore().PeerInfo(h2.ID()))
	h2.Connect(ctx, h1.Peerstore().PeerInfo(h1.ID()))
	fserial := NewFakeSerial()
	// setup some test data
	fserial.Write([]byte("^hello^"))
	bridge, err := NewBridge(ctx, wg, zaptest.NewLogger(t), fserial, Opts{})
	if err != nil {
		t.Fatal(err)
	}
	h1.SetStreamHandler(ProtocolID, bridge.StreamHandler)
	wg.Add(1)
	go func() {
		defer wg.Done()
		s, err := h2.NewStream(ctx, h1.ID(), ProtocolID)
		if err != nil {
			t.Error(err)
		}
		if s == nil {
			panic("fuck")
		}
		defer s.Close()
		s.Write([]byte("^yo dawg this is some test data^"))
	}()
	time.Sleep(time.Second * 10)
	cancel()
	wg.Wait()
}

func newHost(ctx context.Context, t *testing.T, addr string, secret []byte) (host.Host, *dht.IpfsDHT) {
	logger := testutils.NewLogger(t)
	ds := testutils.NewDatastore(t)
	ps := testutils.NewPeerstore(t)
	pk, _, err := crypto.GenerateKeyPair(crypto.ECDSA, 256)
	if err != nil {
		t.Fatal(err)
	}
	maddr, err := multiaddr.NewMultiaddr(addr)
	if err != nil {
		t.Fatal(err)
	}
	ht, dt := testutils.NewLibp2pHostAndDHT(
		ctx, t, logger.Desugar(), ds, ps, pk, []multiaddr.Multiaddr{maddr}, secret,
	)
	return ht, dt
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
	fs.mx.Lock()
	defer fs.mx.Unlock()
	if fs.errNextCall {
		return 0, errors.New("error")
	}
	fs.nextRead = data
	return len(data), nil
}

func (fs *FakeSerial) Available() (int, error) {
	fs.mx.Lock()
	defer fs.mx.Unlock()
	if fs.errNextCall {
		return 0, errors.New("error")
	}
	return len(fs.nextRead), nil
}

func (fs *FakeSerial) Read(data []byte) (int, error) {
	fs.mx.Lock()
	defer fs.mx.Unlock()
	if fs.errNextCall {
		return 0, errors.New("error")
	}
	copy(data, fs.nextRead)
	return len(data), nil
}

func (fs *FakeSerial) Flush() error {
	fs.mx.Lock()
	defer fs.mx.Unlock()
	if fs.errNextCall {
		return errors.New("error")
	}
	return nil
}

func (fs *FakeSerial) Close() error {
	fs.mx.Lock()
	defer fs.mx.Unlock()
	if fs.errNextCall {
		return errors.New("error")
	}
	return nil
}
