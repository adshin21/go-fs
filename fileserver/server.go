package fileserver

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/adshin21/go-fs/peertopeer"
	"github.com/adshin21/go-fs/peertopeer/tcp"
	"github.com/adshin21/go-fs/storage"
)

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int64
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
	tee := io.TeeReader(r, buf)

	// write to disk
	n, err := s.store.Write(key, tee)

	if err != nil {
		return err
	}

	msg := Message{
		Payload: MessageStoreFile{
			Key:  key,
			Size: n,
		},
	}

	msgBuf := new(bytes.Buffer)
	if err := gob.NewEncoder(msgBuf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		peer.Send(msgBuf.Bytes())
	}

	time.Sleep(time.Second * 3)

	for _, peer := range s.peers {
		x, err := io.Copy(peer, buf)
		if err != nil {
			return err
		}
		fmt.Printf("received and written %d bytes to disk\n", x)
	}

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

			if err := s.handleMessage(rpc.From, &m); err != nil {
				log.Printf("FS: error: %s\n", err)
				return
			}
		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) handleMessage(from string, m *Message) error {
	switch msg := m.Payload.(type) {
	case MessageStoreFile:
		log.Printf("received data %+v\n", msg)
		return s.handleMessageStoreFile(from, m.Payload.(MessageStoreFile))
	default:
		return fmt.Errorf("unknown message type %T", msg)
	}
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error {
	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not found in the peer list", from)
	}
	n, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}

	fmt.Printf("written %d bytes to disk\n", n)
	peer.(*tcp.TCPPeer).Wg.Done()
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
	gob.Register(MessageStoreFile{})
}
