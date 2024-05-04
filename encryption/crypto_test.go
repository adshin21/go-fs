package encryption

import (
	"bytes"
	"testing"
)

func TestCopyEncryptionDecryption(t *testing.T) {
	data := "hello world"
	src := bytes.NewReader([]byte(data))
	dst := new(bytes.Buffer)
	key := NewEncrpytionKey()
	_, err := CopyEncrypt(key, src, dst)
	if err != nil {
		t.Error(err)
	}

	out := new(bytes.Buffer)
	_, err = CopyDecrypt(key, dst, out)
	if err != nil {
		t.Error(err)
	}

	if data != out.String() {
		t.Fail()
	}
}
