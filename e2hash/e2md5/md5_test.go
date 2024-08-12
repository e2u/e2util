package e2md5

import (
	"testing"
)

func Test_HeadTailHex(t *testing.T) {
	head := make([]byte, 128)
	tail := make([]byte, 128)
	for i := 0; i < 128; i++ {
		head[i] = 'a'
		tail[i] = 'b'
	}
	r1 := HeadTailHex(append(head, tail...))
	t.Log(string(append(head, tail...)))
	t.Log(r1)

	data := make([]byte, 2048)
	for i := 0; i < 128; i++ {
		data[i] = 'a'
		data[(len(data)-128)+i] = 'b'
	}
	r := HeadTailHex(data)
	t.Log(string(data))
	t.Log(r)

	if r1 != r {
		t.Fatalf("hash hex result mismatch")
	}
}
