package e2crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
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
		// b[i] = encoder[b[i]%byte(len(encoder))]
		b[i] = encoder[int(b[i])%len(encoder)]
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

// TOTPAlgorithm represents the hash algorithm used for TOTP
type TOTPAlgorithm string

const (
	TOTPAlgorithmSHA1   TOTPAlgorithm = "SHA1"
	TOTPAlgorithmSHA256 TOTPAlgorithm = "SHA256"
	TOTPAlgorithmSHA512 TOTPAlgorithm = "SHA512"
)

// TOTPConfig holds configuration for TOTP generation/verification
type TOTPConfig struct {
	Secret    string        // Base32 encoded secret
	Digits    int           // Number of digits (6, 7, or 8)
	Algorithm TOTPAlgorithm // Hash algorithm
	Period    int           // Time step in seconds (usually 30)
}

// DefaultTOTPConfig returns a TOTPConfig with default values
func DefaultTOTPConfig(secret string) TOTPConfig {
	return TOTPConfig{
		Secret:    secret,
		Digits:    6,
		Algorithm: TOTPAlgorithmSHA1,
		Period:    30,
	}
}

// Validate checks if the TOTP configuration is valid
func (c TOTPConfig) Validate() error {
	if c.Secret == "" {
		return errors.New("TOTP secret cannot be empty")
	}
	if c.Digits != 6 && c.Digits != 7 && c.Digits != 8 {
		return errors.New("TOTP digits must be 6, 7, or 8")
	}
	if c.Period <= 0 {
		return errors.New("TOTP period must be positive")
	}
	if c.Algorithm != TOTPAlgorithmSHA1 && c.Algorithm != TOTPAlgorithmSHA256 && c.Algorithm != TOTPAlgorithmSHA512 {
		return errors.New("TOTP algorithm must be SHA1, SHA256, or SHA512")
	}
	return nil
}

// VerifyTOTP verifies a TOTP code against a secret at the given time
// timeStep is typically 30 seconds (standard)
// skew allows for time drift (e.g., 1 = ±30 seconds)
// Deprecated: Use VerifyTOTPWithConfig for more flexibility
// func VerifyTOTP(secret string, code string, t time.Time, timeStep int, skew int) bool {
//	config := DefaultTOTPConfig(secret)
//	config.Period = timeStep
//	return VerifyTOTPWithConfig(config, code, t, skew)
//}

// VerifyTOTPWithConfig verifies a TOTP code using the provided configuration
func VerifyTOTPWithConfig(config TOTPConfig, code string, t time.Time, skew int) bool {
	if err := config.Validate(); err != nil {
		return false
	}
	if len(code) != config.Digits {
		return false
	}
	// Try current and adjacent time steps
	for i := -skew; i <= skew; i++ {
		expected, err := generateTOTPWithConfig(config, t, i)
		if err != nil {
			return false
		}
		if subtle.ConstantTimeCompare([]byte(expected), []byte(code)) == 1 {
			return true
		}
	}
	return false
}

// generateTOTP generates a TOTP code for the given time and offset
// Deprecated: Use generateTOTPWithConfig for more flexibility
// func generateTOTP(secret string, t time.Time, timeStep, stepOffset int) string {
//	config := DefaultTOTPConfig(secret)
//	config.Period = timeStep
//	return generateTOTPWithConfig(config, t, stepOffset)
//}

func calculateCounter(t time.Time, config TOTPConfig, stepOffset int64) (uint64, error) {
	// 1. 先進行基礎有效性檢查
	if config.Period <= 0 {
		return 0, errors.New("TOTP period must be positive")
	}

	// 2. 喺 int64 領域內完成所有核心運算
	// 咁樣可以避免中間過程出現 uint64 溢出或者不正確嘅轉型
	totalInt := (t.Unix() / int64(config.Period)) + stepOffset

	// 3. 最後一步：驗證結果是否符合 uint64 嘅範圍要求
	// 呢個檢查係針對最終結果，而唔係針對中間變量
	if totalInt < 0 {
		return 0, errors.New("calculated counter is negative")
	}

	// 4. 既然已經驗證咗 totalInt >= 0，呢度嘅強制轉換就係絕對安全嘅
	// 使用 // nolint:gosec 嚟告訴後續嘅開發者/工具：我已經考慮過安全性

	return uint64(totalInt), nil
}

// generateTOTPWithConfig generates a TOTP code using the provided configuration
func generateTOTPWithConfig(config TOTPConfig, t time.Time, stepOffset int) (string, error) {
	// Decode base32 secret
	key := base32Decode(config.Secret)
	if key == nil {
		return "", errors.New("TOTP secret cannot be empty")
	}

	// Calculate counter
	// counter := uint64(t.Unix()/int64(config.Period)) + uint64(stepOffset)

	// if config.Period <= 0 {
	//	return "", errors.New("TOTP period must be positive")
	//}
	// divisionResult := t.Unix() / int64(config.Period)
	// if divisionResult < 0 {
	//	return "", errors.New("TOTP period must be positive")
	//}
	//
	//// noline:gosec // G115
	// counter := uint64(divisionResult) + uint64(stepOffset)

	counter, err := calculateCounter(t, config, int64(stepOffset))
	if err != nil {
		return "", err
	}

	// Create HMAC based on algorithm
	var mac hash.Hash
	switch config.Algorithm {
	case TOTPAlgorithmSHA256:
		mac = hmac.New(sha256.New, key)
	case TOTPAlgorithmSHA512:
		mac = hmac.New(sha512.New, key)
	default: // SHA1
		mac = hmac.New(sha1.New, key)
	}
	mac.Write(uint64ToBytes(counter))
	sum := mac.Sum(nil)

	// Dynamic truncation
	offset := sum[len(sum)-1] & 0x0F
	code := ((int(sum[offset])&0x7F)<<24 |
		(int(sum[offset+1])&0xFF)<<16 |
		(int(sum[offset+2])&0xFF)<<8 |
		(int(sum[offset+3]) & 0xFF)) % 1000000

	// Format with correct number of digits
	format := fmt.Sprintf("%%0%dd", config.Digits)
	return fmt.Sprintf(format, code), nil
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
func GenerateTOTP(secret string, t time.Time, timeStep int) (string, error) {
	config := DefaultTOTPConfig(secret)
	config.Period = timeStep
	return generateTOTPWithConfig(config, t, 0)
}

func uint64ToBytes(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
