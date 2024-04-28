package peertopeer

// default communication message
// holds any data sent between transport
// in the given network
type RPC struct {
	From    string
	Payload []byte
	Stream  bool
}
