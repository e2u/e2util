package e2auth

import (
	"crypto/md5"
	"encoding/base64"
	"testing"

	"github.com/e2u/e2util/e2crypto"
	"github.com/e2u/e2util/e2hash"
)

func Test_encryptData(t *testing.T) {
	data := []byte("hello world")
	key := []byte(e2hash.HashHex([]byte("123455"), md5.New))
	enc, err := e2crypto.GCMEncryptData(data, key)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(enc)
	dec, err := e2crypto.GCMDecryptData(enc, key)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(dec))
}

func Test_deriveKey(t *testing.T) {
	key := "password"
	dk, err := e2crypto.DeriveKey([]byte(key), "abcdefg")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(base64.StdEncoding.EncodeToString(dk))
}
