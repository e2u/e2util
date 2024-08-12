package e2slice

import (
	"testing"
)

func Test_HasConsecutiveNumbers(t *testing.T) {
	b1 := HasConsecutiveNumbers([]int{1, 2, 3, 4})
	if !b1 {
		t.Fatal()
	}
}

func Test_Copy(t *testing.T) {

	t.Run("test string", func(t *testing.T) {
		src := []string{"a", "b", "c"}
		dest := Copy(src)
		t.Log(dest)
	})

	t.Run("test bytes", func(t *testing.T) {
		src := []byte{'a', 'b', 'c', 'd'}
		dest := Copy(src)
		t.Log(dest)
	})

}
