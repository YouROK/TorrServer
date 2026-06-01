package urlproxy

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	manet "github.com/multiformats/go-multiaddr/net"
)

type contextKey string

const TargetPeerKey contextKey = "p2p-tunsgo-urlproxy"

type streamConn struct {
	network.Stream
}

func (s *streamConn) LocalAddr() net.Addr {
	addr, _ := manet.ToNetAddr(s.Stream.Conn().LocalMultiaddr())
	return addr
}

func (s *streamConn) RemoteAddr() net.Addr {
	addr, _ := manet.ToNetAddr(s.Stream.Conn().RemoteMultiaddr())
	return addr
}

func (s *streamConn) SetDeadline(t time.Time) error      { return s.Stream.SetDeadline(t) }
func (s *streamConn) SetReadDeadline(t time.Time) error  { return s.Stream.SetReadDeadline(t) }
func (s *streamConn) SetWriteDeadline(t time.Time) error { return s.Stream.SetWriteDeadline(t) }

func NewP2PClient(h host.Host, protoID protocol.ID) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				pID, ok := ctx.Value(TargetPeerKey).(peer.ID)
				if !ok {
					return nil, fmt.Errorf("p2p target peer not specified in context")
				}

				stream, err := h.NewStream(ctx, pID, protoID)
				if err != nil {
					return nil, err
				}

				log.Println("[HTTP] Connecting to p2p:", addr)
				_, err = fmt.Fprintf(stream, "CONNECT %s\n", addr)
				if err != nil {
					stream.Reset()
					return nil, err
				}

				return &streamConn{stream}, nil
			},

			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}
