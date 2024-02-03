package e2hash

import (
	"fmt"
	"hash"
)

func HashHex(data []byte, hashFunc func() hash.Hash) string {
	h := hashFunc()
	h.Write(data)
	return fmt.Sprintf("%x", h.Sum(nil))
}
