package peertopeer

// Interface that represents remoto node
type Peer interface {
	Close() error
}

// To handle connections between nodes
// in the network
// Eg - TCP, UDP, WS ...
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC
}
