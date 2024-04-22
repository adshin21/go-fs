package tcp

import (
	"net"
	"sync"
)

type TCPPeer struct {
	// underlying TCP connection of the peer
	net.Conn

	// true if it is an outboud connection
	outbound bool
	Wg       sync.WaitGroup
}

func NewTCPPerr(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
	}
}

// Send implements the peer interface
func (p *TCPPeer) Send(b []byte) error {
	_, err := p.Conn.Write(b)
	return err
}
