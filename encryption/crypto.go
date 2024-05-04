package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
)

func NewEncrpytionKey() []byte {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		panic(err)
	}
	return key
}

func Hash(key string) string {
	// SHA256 hash algorithm
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

func copyStream(stream cipher.Stream, blcSz int, src io.Reader, dst io.Writer) (int, error) {
	buf := make([]byte, 32*1024)
	bw := blcSz

	for {
		n, err := src.Read(buf)

		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			sz, err := dst.Write(buf[:n])
			if err != nil {
				return 0, err
			}
			bw += sz
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return bw, nil
}

func CopyDecrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}

	iv := make([]byte, block.BlockSize())
	if _, err := src.Read(iv); err != nil {
		return 0, err
	}

	stream := cipher.NewCTR(block, iv)
	nw := block.BlockSize()
	return copyStream(stream, nw, src, dst)

}

func CopyEncrypt(key []byte, src io.Reader, dst io.Writer) (int, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return 0, err
	}
	// 16 bytes
	iv := make([]byte, block.BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return 0, err
	}

	// prepending the iv to the file
	if _, err := dst.Write(iv); err != nil {
		return 0, err
	}

	stream := cipher.NewCTR(block, iv)
	nw := block.BlockSize()
	return copyStream(stream, nw, src, dst)
}
