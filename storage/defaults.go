package storage

import (
	"fmt"
)

var BASE_DIR = "data"

type PathKey struct {
	PathName string
	FileName string
}

func (pk *PathKey) GetFilePath() string {
	return fmt.Sprintf("%s/%s", pk.PathName, pk.FileName)
}

var DefaultGetPathFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		FileName: key,
	}
}
