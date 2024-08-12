package e2crypto

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"math/big"

	"golang.org/x/exp/constraints"
)

var (
	encoder = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
)

// RandomString return a random string , removed  / + - and _
func RandomString(n int) string {
	enc := base64.RawURLEncoding
	src := make([]byte, n)
	if _, err := rand.Read(src); err != nil {
		return ""
	}
	dest := make([]byte, enc.EncodedLen(n))
	enc.Encode(dest, src)
	for idx := 0; idx < len(dest); idx++ {
		if dest[idx] == '-' || dest[idx] == '_' || dest[idx] == '+' || dest[idx] == '/' {
			dest[idx] = encoder[idx%len(encoder)]
		}
	}
	if len(dest) > n {
		return string(dest[:n])
	}
	return string(dest)
}

// RandomBytes 返回随机字节数组
func RandomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return []byte("")
	}
	return b
}

func RandomNumber[T constraints.Integer](min, max T) T {
	nb, err := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
	if err != nil {
		return 0
	}

	result := T(nb.Int64()) + min
	return result
}

func RandomFloat[T constraints.Float](min, max T) T {
	if min > max {
		min, max = max, min
	}
	delta := max - min
	nb, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return 0.0
	}
	randomFraction := T(nb.Int64()) / 1000000.0
	return min + randomFraction*delta
}

func RandomElement[T any](sa []T) (T, error) {
	var zero T
	if len(sa) == 0 {
		return zero, errors.New("slice is empty")
	}
	idx := RandomNumber(0, len(sa)-1)
	return sa[idx], nil
}
