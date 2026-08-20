package e2crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
)

// SignatureAlgorithm represents the hash algorithm used for signing
type SignatureAlgorithm string

const (
	SignatureAlgorithmSHA256 SignatureAlgorithm = "SHA256"
	SignatureAlgorithmSHA384 SignatureAlgorithm = "SHA384"
	SignatureAlgorithmSHA512 SignatureAlgorithm = "SHA512"
)

// SignOptions holds options for signing
type SignOptions struct {
	Algorithm SignatureAlgorithm
}

// DefaultSignOptions returns default signing options (SHA256)
func DefaultSignOptions() SignOptions {
	return SignOptions{
		Algorithm: SignatureAlgorithmSHA256,
	}
}

// getHashFunc returns the hash function for the given algorithm
func getHashFunc(alg SignatureAlgorithm) (hash.Hash, crypto.Hash, error) {
	switch alg {
	case SignatureAlgorithmSHA256:
		return sha256.New(), crypto.SHA256, nil
	case SignatureAlgorithmSHA384:
		return sha512.New384(), crypto.SHA384, nil
	case SignatureAlgorithmSHA512:
		return sha512.New(), crypto.SHA512, nil
	default:
		return nil, 0, fmt.Errorf("unsupported signature algorithm: %s", alg)
	}
}

// Sign signs data using RSA private key with the specified algorithm.
// Returns base64-encoded signature.
func Sign(privateKey *rsa.PrivateKey, data []byte, opts SignOptions) (string, error) {
	if privateKey == nil {
		return "", errors.New("private key cannot be nil")
	}
	if len(data) == 0 {
		return "", errors.New("data cannot be empty")
	}

	hashFunc, hashAlg, err := getHashFunc(opts.Algorithm)
	if err != nil {
		return "", err
	}

	hashFunc.Write(data)
	digest := hashFunc.Sum(nil)

	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, hashAlg, digest)
	if err != nil {
		return "", fmt.Errorf("failed to sign data: %w", err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// Verify verifies an RSA signature against the data using the public key.
// The signature should be base64-encoded.
func Verify(publicKey *rsa.PublicKey, data []byte, signature string, opts SignOptions) error {
	if publicKey == nil {
		return errors.New("public key cannot be nil")
	}
	if len(data) == 0 {
		return errors.New("data cannot be empty")
	}
	if signature == "" {
		return errors.New("signature cannot be empty")
	}

	sigBytes, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	hashFunc, hashAlg, err := getHashFunc(opts.Algorithm)
	if err != nil {
		return err
	}

	hashFunc.Write(data)
	digest := hashFunc.Sum(nil)

	return rsa.VerifyPKCS1v15(publicKey, hashAlg, digest, sigBytes)
}

// SignWithPem signs data using an RSA private key from PEM string.
// Convenience function that parses the key and signs in one step.
func SignWithPem(privateKeyPEM string, data []byte, opts SignOptions) (string, error) {
	privateKey, err := ParseRsaPrivateKeyFromPemString(privateKeyPEM)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}
	return Sign(privateKey, data, opts)
}

// VerifyWithPem verifies a signature using an RSA public key from PEM string.
// Convenience function that parses the key and verifies in one step.
func VerifyWithPem(publicKeyPEM string, data []byte, signature string, opts SignOptions) error {
	publicKey, err := ParseRsaPublicKeyFromPemStr(publicKeyPEM)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}
	return Verify(publicKey, data, signature, opts)
}
