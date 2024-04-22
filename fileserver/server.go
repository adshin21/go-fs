package fileserver

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/adshin21/go-fs/peertopeer"
	"github.com/adshin21/go-fs/peertopeer/tcp"
	"github.com/adshin21/go-fs/storage"
)

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key string
}

// Same like transport options with
// storage options
type FileServerOpts struct {
	Transport      peertopeer.Transport
	StoreOpts      storage.StoreOpts
	BootstrapNodes []string
}

// having options for both transport and storage
type FileServer struct {
	FileServerOpts

	// peer lock
	pl    sync.Mutex
	peers map[string]peertopeer.Peer

	store  *storage.Store
	quitch chan (struct{})
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	// Store the file to disk
	// Replicate on the peers over the network

	buf := new(bytes.Buffer)
	msg := Message{
		Payload: MessageStoreFile{
			Key: key,
		},
	}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if err := peer.Send(buf.Bytes()); err != nil {
			return err
		}
	}

	// time.Sleep(3 * time.Second)

	// payload := []byte("LARGE FILE")

	// for _, peer := range s.peers {
	// 	if err := peer.Send(payload); err != nil {
	// 		return err
	// 	}
	// }

	// buf := new(bytes.Buffer)
	// tee := io.TeeReader(r, buf)

	// // write to disk
	// if err := s.store.Write(key, tee); err != nil {
	// 	return err
	// }

	// // make payload and broadcast
	// p := &DataMessage{
	// 	Key:  key,
	// 	Data: buf.Bytes(),
	// }
	// return s.broadcast(&Message{
	// 	From:    "xyz",
	// 	Payload: p,
	// })
	return nil
}

func NewFileServer(opts FileServerOpts) *FileServer {
	return &FileServer{
		FileServerOpts: opts,
		store:          storage.NewStore(opts.StoreOpts),
		quitch:         make(chan struct{}),
		peers:          make(map[string]peertopeer.Peer),
	}
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}
	s.bootstrapNetwork()
	s.loop()
	return nil
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) OnPeer(p peertopeer.Peer) error {
	s.pl.Lock()
	defer s.pl.Unlock()
	s.peers[p.RemoteAddr().String()] = p
	fmt.Printf("FS: connected with remote %s\n", p.RemoteAddr().String())
	return nil
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("file server stopped due to user quit")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var m Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&m); err != nil {
				fmt.Printf("FS: error: %s\n", err)
				continue
			}

			fmt.Printf("FS: payload: %+v\n", m.Payload)

			peer, ok := s.peers[rpc.From]

			if !ok {
				panic("peer not found in peer map")
			}
			b := make([]byte, 1000)
			if _, err := peer.Read(b); err != nil {
				panic(err)
			}
			fmt.Printf("FS: Large file: %s\n", string(b))
			peer.(*tcp.TCPPeer).Wg.Done()

			// if err := s.handleMessage(&m); err != nil {
			// 	log.Println(err)
			// }
		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) handleMessage(m *Message) error {

	// switch msg := m.Payload.(type) {
	// case *DataMessage:
	// 	log.Printf("received data %+v\n", msg)
	// default:
	// 	return fmt.Errorf("unknown message type %T", msg)
	// }
	return nil
}

func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}
		go func(addrr string) {
			if err := s.Transport.Dial(addrr); err != nil {
				log.Println("FS: dial error: ", err)
			}
		}(addr)
	}
	return nil
}

func (s *FileServer) broadcast(m *Message) error {
	// broadcast over network

	// peer is having reader and writer from net.Conn
	peers := []io.Writer{}
	for _, peer := range s.peers {
		ra := peer.RemoteAddr().String()
		fmt.Printf("Remote addr = %s\n", ra)
		peers = append(peers, peer)
	}

	// fmt.Printf("Total peers are: %+v\n", peers)
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(m)
}

func init() {
	gob.Register(&MessageStoreFile{})
}
