package peertopeer

import "net"

// Interface that represents remoto node
type Peer interface {
	net.Conn
	Send([]byte) error
}

// To handle connections between nodes
// in the network
// Eg - TCP, UDP, WS ...
type Transport interface {
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
