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

type Bridge struct {
	serial          *term.Term
	ctx             context.Context
	cancelFunc      context.CancelFunc
	mx              sync.Mutex
	authorizedPeers map[peer.ID]bool
	logger          *zap.Logger
}

// SerialRead is used to read data off the serial interface
func (b *Bridge) SerialRead(out []byte) (int, error) {
	return b.serial.Read(out)
}

// SerialWrite is used to write data into the serial interface
func (b *Bridge) SerialWrite(data []byte) (int, error) {
	return b.serial.Write(data)
}

// StreamHandler is used to open a bi-directional stream.
// Only one stream may be opened at a time
func (b *Bridge) StreamHandler(stream network.Stream) {
	if !b.authorizedPeers[stream.Conn().RemotePeer()] {
		_, err := stream.Write([]byte("unauthorized"))
		if err != nil {
			b.logger.Error("failed to write response back")
		}
		stream.Close()
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
