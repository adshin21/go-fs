package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func NewEncrpytionKey() []byte {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		panic(err)
	}
	return key
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
	buf := make([]byte, 32*1024)
	stream := cipher.NewCTR(block, iv)
	mw := block.BlockSize()

	for {
		n, err := src.Read(buf)

		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			sz, err := dst.Write(buf[:n])
			if err != nil {
				return 0, err
			}
			mw += sz
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return mw, nil
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

	buf := make([]byte, 32*1024)
	stream := cipher.NewCTR(block, iv)
	mw := block.BlockSize()

	for {
		n, err := src.Read(buf)

		if n > 0 {
			stream.XORKeyStream(buf, buf[:n])
			sz, err := dst.Write(buf[:n])
			if err != nil {
				return 0, err
			}
			mw += sz
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return mw, nil
}
