package e2crypto

import (
	"testing"
)

func TestSignAndVerify(t *testing.T) {
	// Generate a key pair for testing
	privateKey, publicKey, err := GenerateRsaKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	data := []byte("message to sign")
	opts := DefaultSignOptions()

	t.Run("sign and verify with SHA256", func(t *testing.T) {
		signature, err := Sign(privateKey, data, opts)
		if err != nil {
			t.Fatalf("Sign() error = %v", err)
		}

		if signature == "" {
			t.Error("Sign() returned empty signature")
		}

		// Verify with correct public key
		err = Verify(publicKey, data, signature, opts)
		if err != nil {
			t.Errorf("Verify() failed for valid signature: %v", err)
		}

		// Verify with wrong data should fail
		wrongData := []byte("wrong message")
		err = Verify(publicKey, wrongData, signature, opts)
		if err == nil {
			t.Error("Verify() should fail for wrong data")
		}
	})

	t.Run("sign with different algorithms", func(t *testing.T) {
		algorithms := []SignatureAlgorithm{
			SignatureAlgorithmSHA256,
			SignatureAlgorithmSHA384,
			SignatureAlgorithmSHA512,
		}

		for _, alg := range algorithms {
			opts := SignOptions{Algorithm: alg}
			signature, err := Sign(privateKey, data, opts)
			if err != nil {
				t.Errorf("Sign() with %s error = %v", alg, err)
				continue
			}

			err = Verify(publicKey, data, signature, opts)
			if err != nil {
				t.Errorf("Verify() with %s failed: %v", alg, err)
			}
		}
	})

	t.Run("nil key handling", func(t *testing.T) {
		_, err := Sign(nil, data, opts)
		if err == nil {
			t.Error("Sign() should return error for nil private key")
		}

		sig, _ := Sign(privateKey, data, opts)
		err = Verify(nil, data, sig, opts)
		if err == nil {
			t.Error("Verify() should return error for nil public key")
		}
	})

	t.Run("empty data handling", func(t *testing.T) {
		_, err := Sign(privateKey, []byte{}, opts)
		if err == nil {
			t.Error("Sign() should return error for empty data")
		}
	})

	t.Run("invalid signature format", func(t *testing.T) {
		err := Verify(publicKey, data, "invalid-signature", opts)
		if err == nil {
			t.Error("Verify() should return error for invalid signature format")
		}
	})
}

func TestSignWithPem(t *testing.T) {
	// Generate a key pair
	privateKey, _, err := GenerateRsaKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Export to PEM
	privateKeyPEM := ExportRsaPrivateKeyAsPemStr(privateKey)

	data := []byte("message to sign")
	opts := DefaultSignOptions()

	t.Run("sign with PEM", func(t *testing.T) {
		signature, err := SignWithPem(privateKeyPEM, data, opts)
		if err != nil {
			t.Fatalf("SignWithPem() error = %v", err)
		}

		if signature == "" {
			t.Error("SignWithPem() returned empty signature")
		}
	})

	t.Run("sign with invalid PEM", func(t *testing.T) {
		_, err := SignWithPem("invalid-pem", data, opts)
		if err == nil {
			t.Error("SignWithPem() should return error for invalid PEM")
		}
	})
}