package e2hash

import (
	"fmt"
	"hash"
	"io"
)

func HashHex(data []byte, hashFunc func() hash.Hash) string {
	h := hashFunc()
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// HashHexReader computes a hash from an io.Reader using streaming
// This is more memory-efficient than reading the entire content into memory
func HashHexReader(r io.Reader, hashFunc func() hash.Hash) (string, error) {
	h := hashFunc()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
