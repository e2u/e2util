package e2crypto

import (
	"crypto/rand"
	"errors"
	"math/big"

	"golang.org/x/exp/constraints"
)

var (
	encoder = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
)

// RandomString returns a random string of length n or an error
func RandomString(n int) (string, error) {
	if n <= 0 {
		return "", errors.New("length must be positive")
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = encoder[b[i]%byte(len(encoder))]
	}
	return string(b), nil
}

// RandomBytes returns n random bytes or an error
func RandomBytes(n int) ([]byte, error) {
	if n <= 0 {
		return nil, errors.New("length must be positive")
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

func RandomNumber[T constraints.Integer](min, max T) (T, error) {
	if min > max {
		return 0, errors.New("min cannot be greater than max")
	}
	delta := max - min + 1
	nb, err := rand.Int(rand.Reader, big.NewInt(int64(delta)))
	if err != nil {
		return 0, err
	}
	return T(nb.Int64()) + min, nil
}

func RandomFloat[T constraints.Float](min, max T) (T, error) {
	if min > max {
		min, max = max, min
	}
	delta := max - min
	nb, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return 0, err
	}
	randomFraction := T(nb.Int64()) / 1000000.0
	return min + randomFraction*delta, nil
}

func RandomElement[T any](sa []T) (T, error) {
	var zero T
	if len(sa) == 0 {
		return zero, errors.New("slice is empty")
	}
	idx, err := RandomNumber(0, len(sa)-1)
	if err != nil {
		return zero, err
	}
	return sa[idx], nil
}
