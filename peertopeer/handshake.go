package peertopeer

type HandshakeFunc func(Peer) error

func NOPHandshakeFunc(Peer) error {
	return nil
}
