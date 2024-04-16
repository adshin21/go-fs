package storage

import (
	"crypto/sha1"
	"encoding/hex"
	"strings"
)

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
		FileName: hashStr,
	}
}

func cleanPath(s string) string {
	return strings.TrimRight(s, "/") + "/"
}

