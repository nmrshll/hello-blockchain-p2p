package mocknet

import (
	//	"fmt"
	"io"
	"sync"
	"time"

	inet "gx/ipfs/QmSTaEYUgDe1r581hxyd2u9582Hgp3KX4wGwYbRqz2u9Qh/go-libp2p-net"
	peer "gx/ipfs/QmbNepETomvmXfz1X5pHNFD2QuPqnqi47dTd94QJWSorQ3/go-libp2p-peer"
)

// link implements mocknet.Link
// and, for simplicity, inet.Conn
type link struct {
	mock        *mocknet
	nets        []*peernet
	opts        LinkOptions
	ratelimiter *ratelimiter
	// this could have addresses on both sides.

	sync.RWMutex
}

func newLink(mn *mocknet, opts LinkOptions) *link {
	l := &link{mock: mn,
		opts:        opts,
		ratelimiter: NewRatelimiter(opts.Bandwidth)}
	return l
}

func (l *link) newConnPair(dialer *peernet) (*conn, *conn) {
	l.RLock()
	defer l.RUnlock()

	c1 := newConn(l.nets[0], l.nets[1], l, inet.DirOutbound)
	c2 := newConn(l.nets[1], l.nets[0], l, inet.DirInbound)
	c1.rconn = c2
	c2.rconn = c1

	if dialer == c1.net {
		return c1, c2
	}
	return c2, c1
}

func (l *link) newStreamPair() (*stream, *stream) {
	ra, wb := io.Pipe()
	rb, wa := io.Pipe()

	sa := NewStream(wa, ra, inet.DirOutbound)
	sb := NewStream(wb, rb, inet.DirInbound)
	return sa, sb
}

func (l *link) Networks() []inet.Network {
	l.RLock()
	defer l.RUnlock()

	cp := make([]inet.Network, len(l.nets))
	for i, n := range l.nets {
		cp[i] = n
	}
	return cp
}

func (l *link) Peers() []peer.ID {
	l.RLock()
	defer l.RUnlock()

	cp := make([]peer.ID, len(l.nets))
	for i, n := range l.nets {
		cp[i] = n.peer
	}
	return cp
}

func (l *link) SetOptions(o LinkOptions) {
	l.opts = o
	l.ratelimiter.UpdateBandwidth(l.opts.Bandwidth)
}

func (l *link) Options() LinkOptions {
	return l.opts
}

func (l *link) GetLatency() time.Duration {
	return l.opts.Latency
}

func (l *link) RateLimit(dataSize int) time.Duration {
	return l.ratelimiter.Limit(dataSize)
}
