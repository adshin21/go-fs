package storage

type GetPathFunc func(string) string

var DefaultGetPathFunc = func(key string) string {
	return key
}
