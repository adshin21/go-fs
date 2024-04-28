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
	"github.com/adshin21/go-fs/storage"
)

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	Key  string
	Size int64
}

type MessageGetFile struct {
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

func (s *FileServer) Get(key string) (io.Reader, error) {
	if s.store.Has(key) {
		return s.store.Read(key)
	}
	fmt.Printf("don't have file [%s] locally\n", key)
	msg := Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}
	if err := s.broadcast(&msg); err != nil {
		return nil, err
	}

	for _, peer := range s.peers {
		buf := new(bytes.Buffer)
		n, err := io.Copy(buf, peer)
		if err != nil {
			return nil, err
		}
		fmt.Printf("FS: received bytes over the network: %d\n", n)
		fmt.Println(buf.String())
	}
	select {}
	return nil, nil
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

	if err := s.broadcast(&msg); err != nil {
		return err
	}

	time.Sleep(time.Second * 3)

	for _, peer := range s.peers {
		peer.Send([]byte{peertopeer.IncomingStream})
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
		log.Println("file server stopped due to error or user quit action")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var m Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&m); err != nil {
				fmt.Printf("FS: decoding errror: %s\n", err)
			}
			fmt.Printf("FS: payload: %+v\n", m.Payload)

			if err := s.handleMessage(rpc.From, &m); err != nil {
				log.Printf("FS: handle message error: %s\n", err)
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
	case MessageGetFile:
		log.Printf("recieved instruction to fetch file %+v\n", msg)
		return s.handleMessageGetFile(from, m.Payload.(MessageGetFile))
	default:
		return fmt.Errorf("unknown message type %T", msg)
	}
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile) error {
	if !s.store.Has(msg.Key) {
		return fmt.Errorf("FS: file does not exists [%s] locally", msg.Key)
	}
	r, err := s.store.Read(msg.Key)
	if err != nil {
		return err
	}

	peer, ok := s.peers[from]
	if !ok {
		return fmt.Errorf("peer %s not found", from)
	}

	n, err := io.Copy(peer, r)
	if err != nil {
		return err
	}
	fmt.Printf("FS: written %d byte over the network to %s\n", n, from)
	return nil
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

	fmt.Printf("FS: [%s] written %d bytes to disk\n", s.Transport.Addr(), n)
	peer.CloseStream()
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

func (s *FileServer) broadcast(msg *Message) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}
	for _, peer := range s.peers {
		peer.Send([]byte{peertopeer.IncomingMessage})
		peer.Send(buf.Bytes())
	}
	return nil
}

func (s *FileServer) stream(m *Message) error {
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
	gob.Register(MessageGetFile{})
}
