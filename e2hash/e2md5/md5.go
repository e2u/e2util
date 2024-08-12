package e2md5

import (
	"bytes"
	"crypto/md5"
	"fmt"

	"github.com/e2u/e2util/e2exec"
)

func MD5HexString(v []byte) string {
	return fmt.Sprintf("%x", md5.Sum(v))
}

func HeadTailHex(v []byte) string {
	if len(v) <= 1024 {
		return fmt.Sprintf("%x", md5.Sum(v))
	}
	combined := make([]byte, 256)
	reader := bytes.NewReader(v)
	e2exec.Must(reader.Read(combined[0:128]))
	e2exec.Must(reader.ReadAt(combined[128:], int64(len(v)-128)))
	return fmt.Sprintf("%x", md5.Sum(combined))
}
