package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	"github.com/adshin21/go-fs/encryption"
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
		EncKey:    encryption.NewEncrpytionKey(),
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
	s2 := getServer(":4000")
	s3 := getServer(":1122", ":3000", ":4000")

	go func() {
		go func() { log.Fatal(s1.Start()) }()
		time.Sleep(time.Second * 1)
		go func() { log.Fatal(s2.Start()) }()
	}()

	time.Sleep(time.Second * 5)

	go s3.Start()
	time.Sleep(time.Second * 5)

	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("testdata_%d", i)
		data := bytes.NewReader([]byte(fmt.Sprintf("My test data %d", i)))
		s3.StoreData(key, data)

		// time.Sleep(time.Second * 15)
		// Go an remove the file from s3 as delete
		// logic to be implemented
		// r, err := s3.Get(key)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// b, err := io.ReadAll(r)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// fmt.Println(string(b))
	}

	// select {}

}
