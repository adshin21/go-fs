package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCASPathTransformFunc(t *testing.T) {
	key := "abc"
	exp := "a9993e364706816aba3e25717850c26c9cd0d89d"
	val := CASPathTransformFunc(key)
	assert.Equal(t, val.FileName, exp)
}
