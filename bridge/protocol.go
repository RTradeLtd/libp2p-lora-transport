package bridge

import (
	"bufio"
	"context"
	"io"
	"sync"

	"github.com/libp2p/go-libp2p-core/network"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/pkg/term"
	"go.uber.org/zap"
)

// ProtocolID is the protocol ID of the liblora bridge
var ProtocolID = protocol.ID("/liblora-bridge/0.0.1")

// Bridge allows authorized peers to open a stream
// and read/write data through the LoRa bridge
type Bridge struct {
	serial          *term.Term
	ctx             context.Context
	mx              sync.Mutex
	authorizedPeers map[peer.ID]bool
	logger          *zap.Logger
}

// Opts allows configuring the bridge
type Opts struct {
	SerialDevice    string
	Baud            int
	AuthorizedPeers map[peer.ID]bool // empty means allow all
}

// NewBridge returns an initialized bridge, suitable for use a LibP2P protocol
func NewBridge(ctx context.Context, logger *zap.Logger, opt Opts) (*Bridge, error) {
	trm, err := term.Open(opt.SerialDevice, term.Speed(opt.Baud))
	if err != nil {
		return nil, err
	}
	return &Bridge{
		serial:          trm,
		ctx:             ctx,
		authorizedPeers: opt.AuthorizedPeers,
		logger:          logger.Named("lora.bridge"),
	}, nil
}

// StreamHandler is used to open a bi-directional stream.
// Only one stream may be opened at a time
func (b *Bridge) StreamHandler(stream network.Stream) {
	defer stream.Reset()
	if b.authorizedPeers != nil && !b.authorizedPeers[stream.Conn().RemotePeer()] {
		_, err := stream.Write([]byte("unauthorized"))
		if err != nil {
			b.logger.Error("failed to write response back")
		}
	}
	b.mx.Lock()
	defer b.mx.Unlock()
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-b.ctx.Done():
				return
			default:
				size, err := b.serial.Available()
				if err != nil && err != io.EOF {
					b.logger.Error("error getting available serial data", zap.Error(err))
					return
				}
				if size == 0 {
					continue
				}
				data := make([]byte, size)
				s, err := b.serial.Read(data)
				if err != nil && err != io.EOF {
					b.logger.Error("error reading serial data", zap.Error(err))
					return
				}
				// skip improperly formatted messages
				if data[0] != '^' || data[len(data)-1] != '^' {
					continue
				}
				_, err = stream.Write(data[:s])
				if err != nil {
					b.logger.Error("failed to write serial data", zap.Error(err))
					return
				}
			}
		}
	}()
	go func() {
		defer wg.Done()
		reader := bufio.NewReader(stream)
		for {
			select {
			case <-b.ctx.Done():
				return
			default:
				if reader.Buffered() > 0 {
					data := make([]byte, reader.Buffered())
					_, err := reader.Read(data)
					if err != nil {
						b.logger.Error("failed to read data", zap.Error(err))
						return
					}
					_, err = b.serial.Write(append(data, '\n'))
					if err != nil {
						b.logger.Error("failed to write into serial interface", zap.Error(err))
						return
					}
				}
			}
		}
	}()
	wg.Wait()

}
