package e2crypto

import (
	"crypto/rand"
	"fmt"
	"strings"
	"testing"
)

func Test_RandomNumber(t *testing.T) {
	t.Run("01", func(t *testing.T) {
		mi, ma := 10, 100
		r := RandomNumber(mi, ma)
		if r < mi || r > ma {
			t.Errorf("Random number should be between %d and %d", mi, ma)
		}
		t.Log(r)
	})

	t.Run("02", func(t *testing.T) {
		mi, ma := 10, 100
		for i := 0; i < 100; i++ {
			r := RandomNumber(mi, ma)
			if r < mi || r > ma {
				t.Errorf("Random number should be between %d and %d", mi, ma)
			}
			t.Log(r)
		}
	})

	t.Run("03", func(t *testing.T) {
		mi, ma := -10, 10
		for i := 0; i < 100; i++ {
			r := RandomNumber(mi, ma)
			if r < mi || r > ma {
				t.Errorf("Random number should be between %d and %d", mi, ma)
			}
			t.Log(r)
		}
	})
}

func Test_RandomFloat(t *testing.T) {
	t.Run("01", func(t *testing.T) {
		mi, ma := 10.0, 100.0
		r := RandomFloat(mi, ma)
		if r < 0 || r > ma {
			t.Errorf("Random number should be between %f and %f", mi, ma)
		}
		t.Log(r)
	})

	t.Run("02", func(t *testing.T) {
		mi, ma := 10.0, 100.0
		for i := 0; i < 100; i++ {
			r := RandomFloat(mi, ma)
			if r < mi || r > ma {
				t.Errorf("Random number should be between %f and %f", mi, ma)
			}
			t.Log(r)
		}
	})

	t.Run("03", func(t *testing.T) {
		mi, ma := -10.0, 10.0
		for i := 0; i < 100; i++ {
			r := RandomFloat(mi, ma)
			if r < mi || r > ma {
				t.Errorf("Random number should be between %f and %f", mi, ma)
			}
			t.Log(r)
		}
	})
}

func Test_RandomPrime(t *testing.T) {
	bi, err := rand.Prime(rand.Reader, 24)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(bi)
}

func Test_RandomString(t *testing.T) {
	for i := 0; i < 1000; i++ {
		r := RandomString(8)
		if strings.Index(r, "+") >= 0 ||
			strings.Index(r, "/") >= 0 ||
			strings.Index(r, "-") >= 0 ||
			strings.Index(r, "_") >= 0 {
			t.Fatal(r)
		}
		t.Log(r)
	}
}
