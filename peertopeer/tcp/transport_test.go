package tcp

import (
	"testing"

	"github.com/adshin21/go-fs/peertopeer"
	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	opts := TCPTransportOpts{
		ListenAdrr:    ":3000",
		HandshakeFunc: peertopeer.NOPHandshakeFunc,
		Decoder:       peertopeer.DefaultDecoder{},
	}
	tr := NewTCPTransport(opts)
	assert.Equal(t, tr.ListenAdrr, ":3000")
	assert.Nil(t, tr.ListenAndAccept())
}
