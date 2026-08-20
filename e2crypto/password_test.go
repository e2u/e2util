package e2crypto

import (
	"strings"
	"testing"
)

// Test Password Hashing
func TestHashPassword(t *testing.T) {
	t.Run("hash and verify", func(t *testing.T) {
		password := "securePassword123!"
		hash, err := HashPassword(password)
		if err != nil {
			t.Fatalf("HashPassword() error = %v", err)
		}

		// Verify correct password
		if !VerifyPassword(password, hash) {
			t.Error("VerifyPassword() should return true for correct password")
		}

		// Verify wrong password fails
		if VerifyPassword("wrongpassword", hash) {
			t.Error("VerifyPassword() should return false for wrong password")
		}
	})

	t.Run("password too short", func(t *testing.T) {
		_, err := HashPassword("short")
		if err == nil {
			t.Error("HashPassword() should return error for short password")
		}
	})

	t.Run("different passwords produce different hashes", func(t *testing.T) {
		password1 := "password123!"
		password2 := "password123!"

		hash1, _ := HashPassword(password1)
		hash2, _ := HashPassword(password2)

		// Same password should produce different hashes (due to salt)
		if hash1 == hash2 {
			t.Error("Same password should produce different hashes due to salt")
		}
	})
}

// Test HMAC
func TestHMAC(t *testing.T) {
	key := []byte("secret-key")
	data := []byte("message to authenticate")

	t.Run("HMAC-SHA256", func(t *testing.T) {
		mac := HMACSHA256(key, data)
		if len(mac) != 64 { // SHA256 hex = 64 chars
			t.Errorf("HMACSHA256() length = %d, want 64", len(mac))
		}

		// Same input should produce same HMAC
		mac2 := HMACSHA256(key, data)
		if mac != mac2 {
			t.Error("HMAC should be deterministic")
		}

		// Different key should produce different HMAC
		differentKey := []byte("different-key")
		mac3 := HMACSHA256(differentKey, data)
		if mac == mac3 {
			t.Error("Different key should produce different HMAC")
		}

		// Verify HMAC
		if !VerifyHMAC(key, data, mac) {
			t.Error("VerifyHMAC should return true for valid MAC")
		}

		// Verify with wrong MAC should fail
		if VerifyHMAC(key, data, "wrongmac") {
			t.Error("VerifyHMAC should return false for invalid MAC")
		}
	})

	t.Run("HMAC-SHA512", func(t *testing.T) {
		mac := HMACSHA512(key, data)
		if len(mac) != 128 { // SHA512 hex = 128 chars
			t.Errorf("HMACSHA512() length = %d, want 128", len(mac))
		}
	})

	t.Run("HMAC is case sensitive", func(t *testing.T) {
		mac := HMACSHA256(key, data)
		upperMAC := strings.ToUpper(mac)

		if mac == upperMAC {
			t.Skip("MAC is already uppercase")
		}

		if VerifyHMAC(key, data, upperMAC) {
			t.Error("HMAC verification should be case sensitive")
		}
	})
}
