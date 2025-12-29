package e2strconv

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_MustParseInt64(t *testing.T) {
	MustParseInt64("abc", 10, 64)
	numbers := []int{-1, -1, -1, -1, -1, -1}
	fmt.Println(numbers)
}

func Test_ParseInt(t *testing.T) {
	t.Run("valid int8", func(t *testing.T) {
		v, err := ParseInt[int8]("127")
		assert.NoError(t, err)
		assert.Equal(t, int8(127), v)
	})

	t.Run("invalid int8", func(t *testing.T) {
		_, err := ParseInt[int8]("128")
		assert.Error(t, err)

	})

	t.Run("valid int16", func(t *testing.T) {
		v, err := ParseInt[int16]("128")
		assert.NoError(t, err)
		assert.Equal(t, int16(128), v)
	})

	t.Run("invalid int16", func(t *testing.T) {
		_, err := ParseInt[int16]("32768")
		assert.Error(t, err)
	})

	t.Run("valid int32", func(t *testing.T) {
		v, err := ParseInt[int32]("32768")
		assert.NoError(t, err)
		assert.Equal(t, int32(32768), v)
	})

	t.Run("invalid int32", func(t *testing.T) {
		_, err := ParseInt[int32]("2147483648")
		if assert.Error(t, err) {
			t.Log(err)
		}
	})

	t.Run("valid int64", func(t *testing.T) {
		v, err := ParseInt[int64]("2147483648")
		assert.NoError(t, err)
		assert.Equal(t, int64(2147483648), v)
	})

	t.Run("invalid int64", func(t *testing.T) {
		_, err := ParseInt[int64]("9223372036854775808")
		if assert.Error(t, err) {
			t.Log(err)
		}
	})

	t.Run("valid int", func(t *testing.T) {
		v, err := ParseInt[int]("100001")
		assert.NoError(t, err)
		assert.Equal(t, 100001, v)
	})

	t.Run("invalid int", func(t *testing.T) {
		_, err := ParseInt[int]("9223372036854775808")
		if assert.Error(t, err) {
			t.Log(err)
		}
	})

}
