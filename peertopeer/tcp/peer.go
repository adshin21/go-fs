package tcp

import "net"

type TCPPeer struct {
	conn net.Conn

	// true if it is an outboud connection
	outbound bool
}

func NewTCPPerr(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
} 

// Close implements the peer interface
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}
