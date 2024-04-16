package main

import (
	"fmt"
	"log"

	"github.com/adshin21/go-fs/peertopeer"
	"github.com/adshin21/go-fs/peertopeer/tcp"
)

func OnPeer(peer peertopeer.Peer) error {
	peer.Close()
	// fmt.Println("doing something with peer in main.go")
	return nil
}

func main() {

	tcpOpts := tcp.TCPTransportOpts{
		ListenAdrr:    ":3000",
		HandshakeFunc: peertopeer.NOPHandshakeFunc,
		Decoder:       peertopeer.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	tr := tcp.NewTCPTransport(tcpOpts)

	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("Msg: %+v\n", msg)
		}
	}()

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
