package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	dopts "github.com/libp2p/go-libp2p-kad-dht/opts"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"
	"github.com/pkg/term"

	bridge "github.com/RTradeLtd/libp2p-lora-transport/bridge"
	datastore "github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	"github.com/ipfs/go-ipns"
	"github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	host "github.com/libp2p/go-libp2p-core/host"
	peerstore "github.com/libp2p/go-libp2p-core/peerstore"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	record "github.com/libp2p/go-libp2p-record"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

var (
	addr   = flag.String("address", "/ip4/0.0.0.0/tcp/4005", "host multi address")
	device = flag.String("device", "/dev/ttyACM0", "serial device name")
	baud   = flag.Int("baud", 2500000, "default serial device baud")
)

func init() {
	flag.Parse()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := &sync.WaitGroup{}
	ds := dssync.MutexWrap(datastore.NewMapDatastore())
	ps := pstoremem.NewPeerstore()
	pk, _, err := crypto.GenerateKeyPair(crypto.ECDSA, 256)
	if err != nil {
		log.Fatal(err)
	}
	logger := zap.NewExample()
	maddr, err := multiaddr.NewMultiaddr(*addr)
	if err != nil {
		log.Fatal(err)
	}
	h, dt, err := newLibp2pHostAndDHT(ctx, logger, ds, ps, pk, []multiaddr.Multiaddr{maddr})
	if err != nil {
		log.Fatal(err)
	}
	_ = dt
	logger.Info("host information", zap.String("peer.id", h.ID().String()), zap.String("address", maddr.String()))
	var doneChan = make(chan bool, 1)
	wg.Add(1)
	go handleExit(ctx, cancel, wg, doneChan)
	// setup the bridge
	trm, err := term.Open(*device, term.Speed(*baud))
	if err != nil {
		log.Fatal(err)
	}
	protocolBridge, err := bridge.NewBridge(ctx, wg, logger, trm, bridge.Opts{})
	if err != nil {
		log.Fatal(err)
	}
	h.SetStreamHandler(bridge.ProtocolID, protocolBridge.StreamHandler)
	// wait
	wg.Wait()
	protocolBridge.Close()
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

func newLibp2pHostAndDHT(
	ctx context.Context,
	logger *zap.Logger,
	ds datastore.Batching,
	ps peerstore.Peerstore,
	pk crypto.PrivKey,
	addrs []multiaddr.Multiaddr) (host.Host, *dht.IpfsDHT, error) {
	var opts []libp2p.Option
	opts = append(opts,
		libp2p.Identity(pk),
		libp2p.ListenAddrs(addrs...),
		libp2p.Peerstore(ps),
		libp2p.DefaultMuxers,
		libp2p.DefaultTransports,
		libp2p.DefaultSecurity)
	h, err := libp2p.New(ctx, opts...)
	if err != nil {
		return nil, nil, err
	}

	idht, err := dht.New(ctx, h,
		dopts.Validator(record.NamespacedValidator{
			"pk":   record.PublicKeyValidator{},
			"ipns": ipns.Validator{KeyBook: ps},
		}),
	)
	if err != nil {
		return nil, nil, err
	}
	rHost := routedhost.Wrap(h, idht)
	return rHost, idht, nil
}
