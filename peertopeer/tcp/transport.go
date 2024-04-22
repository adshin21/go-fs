package tcp

import (
	"errors"
	"fmt"
	"log"
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

func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.handleConn(conn, true)
	return nil
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
	log.Printf("Accepting connection on %s\n", t.ListenAdrr)
	return nil
}

func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()

		if errors.Is(err, net.ErrClosed) {
			log.Println("TCP: Connection Error: ", err)
			return
		}
		if err != nil {
			fmt.Printf("TCP: Accept Error: %s\n", err)
		}
		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	var err error

	defer func() {
		if err != nil {
			fmt.Printf("TCP: dropping peer connection: %s\n", err)
			conn.Close()
		}
	}()

	peer := NewTCPPerr(conn, outbound)
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
		rpc.From = conn.RemoteAddr().String()
		peer.Wg.Add(1)
		log.Printf("Waiting currently ................")
		t.rpcch <- rpc
		peer.Wg.Wait()
		log.Printf("Waiting finished, continuing...")

	}
}
