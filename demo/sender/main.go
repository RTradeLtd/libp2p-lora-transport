package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gitea.com/lunny/log"
	dopts "github.com/libp2p/go-libp2p-kad-dht/opts"
	routedhost "github.com/libp2p/go-libp2p/p2p/host/routed"

	bridge "github.com/RTradeLtd/libp2p-lora-transport/bridge"
	datastore "github.com/ipfs/go-datastore"
	dssync "github.com/ipfs/go-datastore/sync"
	"github.com/ipfs/go-ipns"
	"github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	host "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	peerstore "github.com/libp2p/go-libp2p-core/peerstore"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p-peerstore/pstoremem"
	record "github.com/libp2p/go-libp2p-record"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

var (
	addr        = flag.String("address", "/ip4/0.0.0.0/tcp/4006", "host multi address")
	peerAddress = flag.String("peer.address", "/ip4/127.0.0.1/tcp/4005", "remote host address")
	peerID      = flag.String("peer.id", "", "peerid of the peer")
	device      = flag.String("device", "/dev/ttyACM0", "serial device name")
	baud        = flag.Int("baud", 2500000, "default serial device baud")
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
	pk, _, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
	if err != nil {
		panic(err)
	}
	logger := zap.NewExample()
	maddr, err := multiaddr.NewMultiaddr(*addr)
	if err != nil {
		panic(err)
	}
	h, dt, err := newLibp2pHostAndDHT(ctx, logger, ds, ps, pk, []multiaddr.Multiaddr{maddr})
	if err != nil {
		panic(err)
	}
	_ = dt
	logger.Info("host information", zap.String("peer.id", h.ID().String()), zap.String("address", maddr.String()))
	var doneChan = make(chan bool, 1)
	wg.Add(1)
	go handleExit(ctx, cancel, wg, doneChan)
	peerAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("%s/p2p/%s", *peerAddress, *peerID))
	if err != nil {
		panic(err)
	}
	pid, err := peer.IDB58Decode(*peerID)
	if err != nil {
		panic(err)
	}
	h.Peerstore().AddAddr(pid, peerAddr, time.Hour)
	peerInfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
	if err != nil {
		panic(err)
	}
	if err := h.Connect(ctx, *peerInfo); err != nil {
		panic(err)
	}
	strm, err := h.NewStream(ctx, pid, bridge.ProtocolID)
	if err != nil {
		panic(err)
	}
	func() {
		defer strm.Reset()
		streamReader := bufio.NewReader(strm)
		inputReader := bufio.NewReader(os.Stdin)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if streamReader.Size() > 0 {
					data := make([]byte, streamReader.Size())
					s, err := streamReader.Read(data)
					if err != nil {
						log.Error(err)
						return
					}
					fmt.Println(string(data[:s]))
				}
				if inputReader.Size() > 0 {
					data, err := inputReader.ReadString('\n')
					if err != nil && err != io.EOF {
						log.Error(err)
						return
					}
					_, err = strm.Write([]byte(data))
					if err != nil {
						log.Error(err)
						return
					}
				}
			}
		}
	}()
	// wait
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
