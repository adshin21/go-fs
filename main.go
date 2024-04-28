package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/adshin21/go-fs/fileserver"
	"github.com/adshin21/go-fs/peertopeer"
	"github.com/adshin21/go-fs/peertopeer/tcp"
	"github.com/adshin21/go-fs/storage"
)

func OnPeer(peer peertopeer.Peer) error {
	// peer.Close()
	// fmt.Println("doing something with peer in main.go")
	return nil
}

func getServer(addr string, nodes ...string) *fileserver.FileServer {
	tcpOpts := tcp.TCPTransportOpts{
		ListenAdrr:    addr,
		HandshakeFunc: peertopeer.NOPHandshakeFunc,
		Decoder:       peertopeer.DefaultDecoder{},
		OnPeer:        OnPeer,
	}
	tr := tcp.NewTCPTransport(tcpOpts)

	fileServerOpts := fileserver.FileServerOpts{
		Transport: tr,
		StoreOpts: storage.StoreOpts{
			BaseDir:     addr + "__",
			GetPathFunc: storage.CASPathTransformFunc,
		},
		BootstrapNodes: nodes,
	}

	s := fileserver.NewFileServer(fileServerOpts)
	tr.OnPeer = s.OnPeer
	return s
}

func main() {

	s1 := getServer(":3000")
	s2 := getServer(":4000", ":3000")

	go func() {
		log.Fatal(s1.Start())
	}()

	time.Sleep(time.Second * 2)

	go s2.Start()
	time.Sleep(time.Second * 1)

	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("testdata_%d", i)
		data := bytes.NewReader([]byte(fmt.Sprintf("My test data %d", i)))
		s2.StoreData(key, data)
	}
	// r, err := s2.Get("testdata")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// b, err := io.ReadAll(r)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Println(string(b))
	select {}
}
