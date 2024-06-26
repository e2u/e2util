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

	head := make([]byte, 128)
	tail := make([]byte, 128)

	reader := bytes.NewReader(v)
	e2exec.Must(reader.Read(head))
	e2exec.Must(reader.ReadAt(tail, int64(len(v)-128)))

	combined := append(head, tail...)
	return fmt.Sprintf("%x", md5.Sum(combined))
}
