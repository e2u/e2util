package e2md5

import (
	"crypto/md5"
	"fmt"
)

func MD5HexString(v []byte) string {
	return fmt.Sprintf("%x", md5.Sum(v))
}
