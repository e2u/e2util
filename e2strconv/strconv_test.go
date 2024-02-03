package e2strconv

import (
	"fmt"
	"testing"
)

func Test_MustParseInt64(t *testing.T) {
	MustParseInt64("abc", 10, 64)
	numbers := []int{-1, -1, -1, -1, -1, -1}
	fmt.Println(numbers)
}
