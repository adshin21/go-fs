package storage

import (
	"io"
	"log"
	"os"
)

type StoreOpts struct {
	GetPathFunc func(string) PathKey
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) writeStream(key string, r io.Reader) error {
	pathKey := s.GetPathFunc(key)

	if err := os.MkdirAll(pathKey.PathName, os.ModePerm); err != nil {
		return err
	}

	comPath := pathKey.FileName()
	f, err := os.Create(comPath)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("STORAGE: written (%d) bytes to disk: %s", n, comPath)
	return nil
}
