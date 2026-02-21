package e2crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"io"
	"math/big"
	"strings"
	"time"

	"golang.org/x/crypto/hkdf"
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

func MustRandomString(n int) string {
	s, _ := RandomString(n)
	return s
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

func MustRandomBytes(n int) []byte {
	b, _ := RandomBytes(n)
	return b
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

func MustRandomNumber[T constraints.Integer](min, max T) T {
	t, _ := RandomNumber[T](min, max)
	return t
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
func MustRandomFloat[T constraints.Float](min, max T) T {
	t, _ := RandomFloat[T](min, max)
	return t
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

func MustRandomElement[T any](sa []T) T {
	t, _ := RandomElement(sa)
	return t
}

func DeriveKey(inputKey []byte, info string) ([]byte, error) {
	df := hkdf.New(sha256.New, inputKey, nil, []byte(info))
	key := make([]byte, 32)
	if _, err := io.ReadFull(df, key); err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}
	return key, nil
}

// TOTP related functions for MFA

// GenerateTOTPSecret generates a new TOTP secret (base32 encoded)
func GenerateTOTPSecret() (string, error) {
	// RFC 4226 recommends 160 bits (20 bytes) minimum
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}
	// Encode to base32 (standard for TOTP)
	return base32Encode(secret), nil
}

// base32Encode encodes bytes to base32 without padding
func base32Encode(data []byte) string {
	const base32Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	result := make([]byte, 0, len(data)*8/5+1)
	for i := 0; i < len(data); i += 5 {
		chunk := data[i:min(i+5, len(data))]
		// Pad with zeros if needed
		padded := make([]byte, 5)
		copy(padded, chunk)

		// Convert 5 bytes to 8 base32 characters
		result = append(result, base32Alphabet[padded[0]>>3])
		result = append(result, base32Alphabet[((padded[0]&0x07)<<2)|(padded[1]>>6)])
		result = append(result, base32Alphabet[((padded[1]&0x3E)>>1)])
		result = append(result, base32Alphabet[((padded[1]&0x01)<<4)|(padded[2]>>4)])
		result = append(result, base32Alphabet[((padded[2]&0x0F)<<1)|(padded[3]>>7)])
		result = append(result, base32Alphabet[((padded[3]&0x7C)>>2)])
		result = append(result, base32Alphabet[((padded[3]&0x03)<<3)|(padded[4]>>5)])
		result = append(result, base32Alphabet[padded[4]&0x1F])
	}
	return string(result)
}

// VerifyTOTP verifies a TOTP code against a secret at the given time
// timeStep is typically 30 seconds (standard)
// skew allows for time drift (e.g., 1 = ±30 seconds)
func VerifyTOTP(secret string, code string, t time.Time, timeStep int, skew int) bool {
	if len(code) != 6 {
		return false
	}
	// Try current and adjacent time steps
	for i := -skew; i <= skew; i++ {
		expected := generateTOTP(secret, t, timeStep, i)
		if subtle.ConstantTimeCompare([]byte(expected), []byte(code)) == 1 {
			return true
		}
	}
	return false
}

// generateTOTP generates a TOTP code for the given time and offset
func generateTOTP(secret string, t time.Time, timeStep, stepOffset int) string {
	// Decode base32 secret
	key := base32Decode(secret)
	if key == nil {
		return ""
	}

	// Calculate counter
	counter := uint64(t.Unix()/int64(timeStep)) + uint64(stepOffset)

	// Create HMAC-SHA1
	mac := hmac.New(sha1.New, key)
	mac.Write(uint64ToBytes(counter))
	hash := mac.Sum(nil)

	// Dynamic truncation
	offset := hash[len(hash)-1] & 0x0F
	code := ((int(hash[offset])&0x7F)<<24 |
		(int(hash[offset+1])&0xFF)<<16 |
		(int(hash[offset+2])&0xFF)<<8 |
		(int(hash[offset+3]) & 0xFF)) % 1000000

	return fmt.Sprintf("%06d", code)
}

// base32Decode decodes a base32 string (case insensitive)
func base32Decode(s string) []byte {
	const base32Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	upper := strings.ToUpper(s)
	result := make([]byte, 0, len(upper)*5/8)
	buffer := uint32(0)
	bitsLeft := 0

	for i := 0; i < len(upper); i++ {
		c := upper[i]
		val := byte(0)
		found := false
		for j := range len(base32Alphabet) {
			if base32Alphabet[j] == c {
				val = byte(j)
				found = true
				break
			}
		}
		if !found {
			continue // Skip invalid characters
		}

		buffer = (buffer << 5) | uint32(val)
		bitsLeft += 5

		if bitsLeft >= 8 {
			bitsLeft -= 8
			result = append(result, byte((buffer>>bitsLeft)&0xFF))
		}
	}

	return result
}

// GenerateTOTP generates a TOTP code for the given time
// This is exported for testing purposes
func GenerateTOTP(secret string, t time.Time, timeStep int) string {
	return generateTOTP(secret, t, timeStep, 0)
}

func uint64ToBytes(v uint64) []byte {
	return []byte{
		byte(v >> 56),
		byte(v >> 48),
		byte(v >> 40),
		byte(v >> 32),
		byte(v >> 24),
		byte(v >> 16),
		byte(v >> 8),
		byte(v),
	}
}
