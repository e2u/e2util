package e2array

import (
	"testing"
)

func Test_HasConsecutiveNumbers(t *testing.T) {
	b1 := HasConsecutiveNumbers([]int{1, 2, 3, 4})
	if !b1 {
		t.Fatal()
	}
}
