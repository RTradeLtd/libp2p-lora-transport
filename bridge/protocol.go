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

// ensure term.Term always satisfies this interface
var _ Serial = (*term.Term)(nil)

// Serial is an interface to provide easy testing of the lora protocol
type Serial interface {
	Write([]byte) (int, error)
	Available() (int, error)
	Read([]byte) (int, error)
	Flush() error
	Close() error
}

// Bridge allows authorized peers to open a stream
// and read/write data through the LoRa bridge
type Bridge struct {
	serial          Serial
	ctx             context.Context
	wg              *sync.WaitGroup
	authorizedPeers map[peer.ID]bool
	readChan        chan []byte
	writeChan       chan []byte
	logger          *zap.Logger
}

// Opts allows configuring the bridge
type Opts struct {
	SerialDevice    string
	Baud            int
	AuthorizedPeers map[peer.ID]bool // empty means allow all
}

// NewBridge returns an initialized bridge, suitable for use a LibP2P protocol
func NewBridge(ctx context.Context, wg *sync.WaitGroup, logger *zap.Logger, serial Serial, opt Opts) (*Bridge, error) {
	bridge := &Bridge{
		serial:          serial,
		ctx:             ctx,
		authorizedPeers: opt.AuthorizedPeers,
		logger:          logger.Named("lora.bridge"),
		readChan:        make(chan []byte),
		writeChan:       make(chan []byte),
		wg:              wg,
	}
	bridge.serialDumper()
	return bridge, nil
}

// serialDumper allows any number of libp2p streams to write
// into the LoRa bridge, or read from it. Reads are sent to any
// active streams
func (b *Bridge) serialDumper() {
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		for {
			select {
			case <-b.ctx.Done():
				return
			case data := <-b.writeChan:
				_, err := b.serial.Write(data)
				if err != nil && err != io.EOF {
					b.logger.Error("failed to write into serial interface", zap.Error(err))
					return
				}
			default:
				avail, err := b.serial.Available()
				if err != nil && err != io.EOF {
					b.logger.Error("serial dumper failed to read available bytes", zap.Error(err))
					return
				}
				if avail == 0 {
					continue
				}
				data := make([]byte, avail)
				s, err := b.serial.Read(data)
				if err != nil && err != io.EOF {
					b.logger.Error("error reading serial data", zap.Error(err))
					return
				}
				// skip improperly formatted messages
				if data[0] != '^' || data[len(data)-1] != '^' {
					continue
				}
				select {
				case b.readChan <- data[:s]:
				case <-b.ctx.Done():
				}
			}
		}
	}()
}

// StreamHandler is used to open a bi-directional stream.
func (b *Bridge) StreamHandler(stream network.Stream) {
	defer stream.Reset()
	if b.authorizedPeers != nil && !b.authorizedPeers[stream.Conn().RemotePeer()] {
		_, err := stream.Write([]byte("unauthorized"))
		if err != nil {
			b.logger.Error("failed to write response back")
		}
	}
	reader := bufio.NewReader(stream)
	for {
		select {
		case <-b.ctx.Done():
			return
		case data := <-b.readChan:
			_, err := stream.Write(data)
			if err != nil {
				b.logger.Error("failed to write into stream")
				return
			}
		default:
			if reader.Size() > 0 {
				data := make([]byte, reader.Size())
				s, err := reader.Read(data)
				if err != nil {
					b.logger.Error("failed to read data from stream buffer", zap.Error(err))
					return
				}
				b.writeChan <- data[:s]
			}
		}
	}
}

// Close is used to shutdown the bridge serial interface
func (b *Bridge) Close() error {
	return b.serial.Close()
}
