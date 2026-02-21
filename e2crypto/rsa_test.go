package e2crypto

import (
	"strings"
	"testing"
)

func TestGenerateRsaKeyPair(t *testing.T) {
	t.Run("default key size", func(t *testing.T) {
		privateKey, publicKey, err := GenerateRsaKeyPair(0)
		if err != nil {
			t.Fatalf("GenerateRsaKeyPair() error = %v", err)
		}

		if privateKey == nil {
			t.Error("GenerateRsaKeyPair() returned nil private key")
		}

		if publicKey == nil {
			t.Error("GenerateRsaKeyPair() returned nil public key")
		}

		// Check default size is 3072
		if privateKey.N.BitLen() != 3072 {
			t.Errorf("Expected default key size 3072, got %d", privateKey.N.BitLen())
		}
	})

	t.Run("custom key size", func(t *testing.T) {
		privateKey, publicKey, err := GenerateRsaKeyPair(2048)
		if err != nil {
			t.Fatalf("GenerateRsaKeyPair() error = %v", err)
		}

		if privateKey.N.BitLen() != 2048 {
			t.Errorf("Expected key size 2048, got %d", privateKey.N.BitLen())
		}

		if publicKey.N.BitLen() != 2048 {
			t.Errorf("Expected public key size 2048, got %d", publicKey.N.BitLen())
		}
	})

	t.Run("key size too small", func(t *testing.T) {
		_, _, err := GenerateRsaKeyPair(1024)
		if err == nil {
			t.Error("GenerateRsaKeyPair() should return error for key size < 2048")
		}
	})

	t.Run("large key size", func(t *testing.T) {
		privateKey, _, err := GenerateRsaKeyPair(4096)
		if err != nil {
			t.Fatalf("GenerateRsaKeyPair() error = %v", err)
		}

		if privateKey.N.BitLen() != 4096 {
			t.Errorf("Expected key size 4096, got %d", privateKey.N.BitLen())
		}
	})
}

func TestParseRsaPrivateKeyFromPemString(t *testing.T) {
	privateKey, _, err := GenerateRsaKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	privateKeyPEM := ExportRsaPrivateKeyAsPemStr(privateKey)

	t.Run("parse valid PEM", func(t *testing.T) {
		parsedKey, err := ParseRsaPrivateKeyFromPemString(privateKeyPEM)
		if err != nil {
			t.Fatalf("ParseRsaPrivateKeyFromPemString() error = %v", err)
		}

		if parsedKey == nil {
			t.Error("ParseRsaPrivateKeyFromPemString() returned nil key")
		}
	})

	t.Run("parse invalid PEM", func(t *testing.T) {
		_, err := ParseRsaPrivateKeyFromPemString("invalid-pem-data")
		if err == nil {
			t.Error("ParseRsaPrivateKeyFromPemString() should return error for invalid PEM")
		}
	})

	t.Run("parse empty PEM", func(t *testing.T) {
		_, err := ParseRsaPrivateKeyFromPemString("")
		if err == nil {
			t.Error("ParseRsaPrivateKeyFromPemString() should return error for empty PEM")
		}
	})
}

func TestParseRsaPublicKeyFromPemStr(t *testing.T) {
	_, publicKey, err := GenerateRsaKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	publicKeyPEM, err := ExportRsaPublicKeyAsPemStr(publicKey)
	if err != nil {
		t.Fatalf("Failed to export public key: %v", err)
	}

	t.Run("parse valid PEM", func(t *testing.T) {
		parsedKey, err := ParseRsaPublicKeyFromPemStr(publicKeyPEM)
		if err != nil {
			t.Fatalf("ParseRsaPublicKeyFromPemStr() error = %v", err)
		}

		if parsedKey == nil {
			t.Error("ParseRsaPublicKeyFromPemStr() returned nil key")
		}
	})

	t.Run("parse invalid PEM", func(t *testing.T) {
		_, err := ParseRsaPublicKeyFromPemStr("invalid-pem-data")
		if err == nil {
			t.Error("ParseRsaPublicKeyFromPemStr() should return error for invalid PEM")
		}
	})
}

func TestExportRsaPublicKeyAsPemStr(t *testing.T) {
	_, publicKey, err := GenerateRsaKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	t.Run("export valid public key", func(t *testing.T) {
		pemStr, err := ExportRsaPublicKeyAsPemStr(publicKey)
		if err != nil {
			t.Fatalf("ExportRsaPublicKeyAsPemStr() error = %v", err)
		}

		if pemStr == "" {
			t.Error("ExportRsaPublicKeyAsPemStr() returned empty string")
		}

		if !strings.Contains(pemStr, "BEGIN RSA PUBLIC KEY") {
			t.Error("PEM should contain BEGIN RSA PUBLIC KEY")
		}
	})

	t.Run("export nil public key", func(t *testing.T) {
		_, err := ExportRsaPublicKeyAsPemStr(nil)
		if err == nil {
			t.Error("ExportRsaPublicKeyAsPemStr() should return error for nil key")
		}
	})
}

func TestExportRsaPrivateKeyAsPemStr(t *testing.T) {
	privateKey, _, err := GenerateRsaKeyPair(2048)
	if err != nil {
		t.Fatalf("Failed to generate key: %v", err)
	}

	t.Run("export valid private key", func(t *testing.T) {
		pemStr := ExportRsaPrivateKeyAsPemStr(privateKey)
		if pemStr == "" {
			t.Error("ExportRsaPrivateKeyAsPemStr() returned empty string")
		}

		if !strings.Contains(pemStr, "BEGIN RSA PRIVATE KEY") {
			t.Error("PEM should contain BEGIN RSA PRIVATE KEY")
		}
	})
}