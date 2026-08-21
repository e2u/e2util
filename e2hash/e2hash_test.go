package e2hash

import (
	"crypto/md5"
	"crypto/sha256"
	"strings"
	"testing"
)

func TestHashHex(t *testing.T) {
	got := HashHex([]byte("hello"), md5.New)
	want := "5d41402abc4b2a76b9719d911017c592"
	if got != want {
		t.Errorf("HashHex(md5) = %s, want %s", got, want)
	}

	got = HashHex([]byte("hello"), sha256.New)
	want = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if got != want {
		t.Errorf("HashHex(sha256) = %s, want %s", got, want)
	}
}

func TestHashHexReader(t *testing.T) {
	got, err := HashHexReader(strings.NewReader("hello"), md5.New)
	if err != nil {
		t.Fatal(err)
	}
	if got != HashHex([]byte("hello"), md5.New) {
		t.Errorf("HashHexReader mismatch: %s", got)
	}
}
