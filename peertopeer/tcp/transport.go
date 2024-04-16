package tcp

import (
	"fmt"
	"net"

	"github.com/adshin21/go-fs/peertopeer"
)

type TCPTransportOpts struct {
	ListenAdrr    string
	HandshakeFunc peertopeer.HandshakeFunc
	Decoder       peertopeer.Decoder
	OnPeer        func(peertopeer.Peer) error
}

// TcpListener
type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan peertopeer.RPC
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan peertopeer.RPC),
	}
}

// for reading incoming messages from other peers in the network
func (t *TCPTransport) Consume() <-chan peertopeer.RPC {
	return t.rpcch
}

func (t *TCPTransport) ListenAndAccept() error {

	listener, err := net.Listen("tcp", t.ListenAdrr)
	if err != nil {
		return err
	}
	t.listener = listener

	// Accepting connection
	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP: Accept Error: %s\n", err)
		}
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn) {
	var err error

	defer func() {
		if err != nil {
			fmt.Printf("TCP: dropping peer connection: %s\n", err)
			conn.Close()
		}
	}()

	peer := NewTCPPerr(conn, true)
	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// Read loop
	rpc := peertopeer.RPC{}
	for {
		err = t.Decoder.Decode(conn, &rpc)
		if err != nil {
			return
		}
		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
	}
}
