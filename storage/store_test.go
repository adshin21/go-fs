package storage

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoreDeleteKey(t *testing.T) {
	s := newStore()
	defer tearDown(t, s)
	key := "xyz"
	data := []byte("some random bytes")
	if _, err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}
	if err := s.Delete(key); err != nil {
		t.Error(err)
	}
}

func TestStore(t *testing.T) {
	s := newStore()
	defer tearDown(t, s)
	key := "xyz"
	data := []byte("some random bytes")

	if _, err := s.writeStream(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}
	_, r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}
	b, _ := io.ReadAll(r)
	assert.Equal(t, string(b), string(data))
	assert.True(t, s.Has(key))
}

func newStore() *Store {
	opts := StoreOpts{
		GetPathFunc: CASPathTransformFunc,
	}
	return NewStore(opts)
}

func tearDown(t *testing.T, s *Store) {
	if err := os.RemoveAll(s.BaseDir); err != nil {
		t.Error("Error while claning up store:", err)
	}
}
