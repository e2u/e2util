package e2crypto

import (
	"bytes"
	"testing"
)

func TestGCMEncryptDecrypt(t *testing.T) {
	// Generate a 32-byte key for AES-256
	key, err := RandomBytes(32)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	t.Run("encrypt and decrypt", func(t *testing.T) {
		plaintext := []byte("Hello, World! This is a secret message.")

		// Encrypt
		ciphertext, err := GCMEncryptData(plaintext, key)
		if err != nil {
			t.Fatalf("GCMEncryptData() error = %v", err)
		}

		if ciphertext == "" {
			t.Error("GCMEncryptData() returned empty ciphertext")
		}

		// Decrypt
		decrypted, err := GCMDecryptData(ciphertext, key)
		if err != nil {
			t.Fatalf("GCMDecryptData() error = %v", err)
		}

		if !bytes.Equal(decrypted, plaintext) {
			t.Errorf("Decrypted text doesn't match original: got %s, want %s", decrypted, plaintext)
		}
	})

	t.Run("different plaintexts produce different ciphertexts", func(t *testing.T) {
		plaintext1 := []byte("Message 1")
		plaintext2 := []byte("Message 2")

		ciphertext1, _ := GCMEncryptData(plaintext1, key)
		ciphertext2, _ := GCMEncryptData(plaintext2, key)

		if ciphertext1 == ciphertext2 {
			t.Error("Different plaintexts should produce different ciphertexts")
		}
	})

	t.Run("same plaintext produces different ciphertexts (nonce)", func(t *testing.T) {
		plaintext := []byte("Same message")

		ciphertext1, _ := GCMEncryptData(plaintext, key)
		ciphertext2, _ := GCMEncryptData(plaintext, key)

		if ciphertext1 == ciphertext2 {
			t.Error("Same plaintext encrypted twice should produce different ciphertexts due to nonce")
		}
	})

	t.Run("decrypt with wrong key fails", func(t *testing.T) {
		plaintext := []byte("Secret message")
		ciphertext, _ := GCMEncryptData(plaintext, key)

		// Generate a different key
		wrongKey, _ := RandomBytes(32)

		_, err := GCMDecryptData(ciphertext, wrongKey)
		if err == nil {
			t.Error("GCMDecryptData() should fail with wrong key")
		}
	})

	t.Run("decrypt tampered ciphertext fails", func(t *testing.T) {
		plaintext := []byte("Secret message")
		ciphertext, _ := GCMEncryptData(plaintext, key)

		// Tamper with ciphertext
		tampered := ciphertext[:len(ciphertext)-1] + "X"

		_, err := GCMDecryptData(tampered, key)
		if err == nil {
			t.Error("GCMDecryptData() should fail with tampered ciphertext")
		}
	})

	t.Run("empty plaintext", func(t *testing.T) {
		plaintext := []byte("")
		ciphertext, err := GCMEncryptData(plaintext, key)
		if err != nil {
			t.Fatalf("GCMEncryptData() error = %v", err)
		}

		decrypted, err := GCMDecryptData(ciphertext, key)
		if err != nil {
			t.Fatalf("GCMDecryptData() error = %v", err)
		}

		if !bytes.Equal(decrypted, plaintext) {
			t.Error("Decrypted empty plaintext doesn't match")
		}
	})

	t.Run("large plaintext", func(t *testing.T) {
		// 1MB of data
		plaintext := make([]byte, 1024*1024)
		for i := range plaintext {
			plaintext[i] = byte(i % 256)
		}

		ciphertext, err := GCMEncryptData(plaintext, key)
		if err != nil {
			t.Fatalf("GCMEncryptData() error = %v", err)
		}

		decrypted, err := GCMDecryptData(ciphertext, key)
		if err != nil {
			t.Fatalf("GCMDecryptData() error = %v", err)
		}

		if !bytes.Equal(decrypted, plaintext) {
			t.Error("Decrypted large plaintext doesn't match")
		}
	})
}