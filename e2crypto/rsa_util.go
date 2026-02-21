package e2crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

func ParseRsaPrivateKeyFromPemString(privatePem string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privatePem))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}

	// Try PKCS8 first, then fall back to PKCS1
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		rsaKey, ok := privateKey.(*rsa.PrivateKey)
		if ok {
			return rsaKey, nil
		}
		return nil, errors.New("parsed key is not an RSA private key")
	}

	// Try PKCS1
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	return privateKey.(*rsa.PrivateKey), nil
}

func ParseRsaPublicKeyFromPemStr(publicKeyPEM string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		break // fall through
	}
	return nil, errors.New("key type is not RSA")
}

func ExportRsaPublicKeyAsPemStr(publicKey *rsa.PublicKey) (string, error) {
	if publicKey == nil {
		return "", errors.New("public key cannot be nil")
	}
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", err
	}
	publicKeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: publicKeyBytes,
		},
	)

	return string(publicKeyPem), nil
}

// DefaultRSAKeyBits is the default RSA key length (3072 bits for security and performance balance)
const DefaultRSAKeyBits = 3072

// MinRSAKeyBits is the minimum recommended RSA key length
const MinRSAKeyBits = 2048

// GenerateRsaKeyPair generates an RSA key pair with the specified bit length.
// If bits is 0, it defaults to DefaultRSAKeyBits.
// Returns error if bits is less than MinRSAKeyBits or key generation fails.
func GenerateRsaKeyPair(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	if bits == 0 {
		bits = DefaultRSAKeyBits
	}
	if bits < MinRSAKeyBits {
		return nil, nil, fmt.Errorf("RSA key size must be at least %d bits for security", MinRSAKeyBits)
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}
	return privateKey, &privateKey.PublicKey, nil
}

func ExportRsaPrivateKeyAsPemStr(privateKey *rsa.PrivateKey) string {
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		},
	)
	return string(privateKeyPem)
}
