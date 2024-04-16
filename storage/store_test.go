package storage

import (
	"bytes"
	"testing"
)

func TestStore(t *testing.T) {
	opts := StoreOpts{
		GetPathFunc: CASPathTransformFunc,
	}
	s := NewStore(opts)
	data := bytes.NewReader([]byte("some random bytes"))
	if err := s.writeStream("xyz", data); err != nil {
		t.Error(err)
	}
}
