package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
)

type PathKey struct {
	PathName string
	Original string
}

func (pk *PathKey) FileName() string {
	return fmt.Sprintf("%s/%s", pk.PathName, pk.Original)
}

func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])
	blockSize := 5
	slicelen := len(hashStr) / blockSize
	path := make([]string, slicelen)

	for i := 0; i < slicelen; i++ {
		from, to := i*blockSize, (i+1)*blockSize
		path[i] = hashStr[from:to]
	}

	return PathKey{
		PathName: strings.Join(path, "/"),
		Original: hashStr,
	}
}
