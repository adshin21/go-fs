package peertopeer

import "net"

// default communication message
// holds any data sent between transport
// in the given network
type RPC struct {
	From    net.Addr
	Payload []byte
}
