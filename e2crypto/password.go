package e2crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordHashCost defines the cost parameter for bcrypt
// Higher cost = more secure but slower (default: 10)
const DefaultPasswordHashCost = 10

// MinPasswordLength defines the minimum acceptable password length
const MinPasswordLength = 8

// HashPassword creates a bcrypt hash of the password using the default cost.
// Returns error if password is too short.
func HashPassword(password string) (string, error) {
	if len(password) < MinPasswordLength {
		return "", fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), DefaultPasswordHashCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// HashPasswordWithCost creates a bcrypt hash with a specific cost.
// Cost should be between 4 and 31 (higher is more secure but slower).
func HashPasswordWithCost(password string, cost int) (string, error) {
	if len(password) < MinPasswordLength {
		return "", fmt.Errorf("password must be at least %d characters", MinPasswordLength)
	}
	if cost < 4 || cost > 31 {
		return "", fmt.Errorf("bcrypt cost must be between 4 and 31")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword checks if the provided password matches the stored hash.
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// HMACSHA256 computes HMAC-SHA256 of data using the provided key.
// Returns the result as a hex-encoded string.
func HMACSHA256(key, data []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// HMACSHA512 computes HMAC-SHA512 of data using the provided key.
// Returns the result as a hex-encoded string.
func HMACSHA512(key, data []byte) string {
	mac := hmac.New(sha512.New, key)
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

// VerifyHMAC verifies that the provided MAC matches the computed HMAC-SHA256.
// Use this for message authentication.
func VerifyHMAC(key, data []byte, mac string) bool {
	expectedMAC := HMACSHA256(key, data)
	return hmac.Equal([]byte(mac), []byte(expectedMAC))
}