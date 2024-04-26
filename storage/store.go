package storage

import (
	"bytes"
	"io"
	"log"
	"os"
)

type StoreOpts struct {
	GetPathFunc func(string) PathKey
	BaseDir     string
}

type Store struct {
	StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.GetPathFunc == nil {
		opts.GetPathFunc = DefaultGetPathFunc
	}

	if opts.BaseDir == "" {
		opts.BaseDir = BASE_DIR
	}

	opts.BaseDir = cleanPath(opts.BaseDir)
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.readStream(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)
	return buf, err
}

func (s *Store) Write(key string, r io.Reader) (int64, error) {
	return s.writeStream(key, r)
}

func (s *Store) Delete(key string) error {
	filePath := s.getCompleteFilePath(key)

	defer func() {
		log.Printf("STORAGE: deleted [%s] from disk", filePath)
	}()

	return os.RemoveAll(filePath)
}

func (s *Store) Has(key string) bool {
	filePath := s.getCompleteFilePath(key)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}
	return true
}

func (s *Store) writeStream(key string, r io.Reader) (int64, error) {
	dir := s.getParentDir(key)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return 0, err
	}

	comPath := s.getCompleteFilePath(key)
	f, err := os.Create(comPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	n, err := io.Copy(f, r)
	if err != nil {
		return 0, err
	}

	log.Printf("STORAGE: written (%d) bytes to disk: %s", n, comPath)
	return n, nil
}

func (s *Store) readStream(key string) (io.ReadCloser, error) {
	path := s.getCompleteFilePath(key)
	return os.Open(path)
}

func (s *Store) getCompleteFilePath(key string) string {
	pathKey := s.GetPathFunc(key)
	return s.BaseDir + pathKey.GetFilePath()
}

func (s *Store) getParentDir(key string) string {
	pathKey := s.GetPathFunc(key)
	return s.BaseDir + pathKey.PathName
}
