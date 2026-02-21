package e2jwt

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/e2u/e2util/e2crypto"
	"github.com/e2u/e2util/e2json"
	"github.com/golang-jwt/jwt/v5"
)

func GenerateWithEncryptSubject(subject any, claims *jwt.RegisteredClaims, secretKey []byte) (string, error) {
	ss := e2json.MustToJSONByte(subject)
	aesKey, err := e2crypto.DeriveKey(secretKey, "encryption-aes-key-from-jwt-key")
	if err != nil {
		return "", err
	}
	enc, err := e2crypto.GCMEncryptData(ss, aesKey)
	if err != nil {
		return "", err
	}
	return generate(enc, claims, secretKey)
}
func verifyEncryptSubjectAndClaims[T any](tokenString string, claims jwt.Claims, secretKey []byte) (T, jwt.Claims, error) {
	var zero T
	if tokenString == "" {
		return zero, nil, errors.New("token is empty")
	}
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return zero, nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return zero, nil, errors.New("invalid token")
	}

	regClaims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return zero, nil, errors.New("invalid claims type")
	}

	if regClaims.Subject == "" {
		return zero, nil, errors.New("subject is empty")
	}

	aesKey, err := e2crypto.DeriveKey(secretKey, "encryption-aes-key-from-jwt-key")
	if err != nil {
		return zero, nil, err
	}
	decSub, err := e2crypto.GCMDecryptData(regClaims.Subject, aesKey)

	var result T
	if err = e2json.MustFromJSONByte(decSub, &result); err != nil {
		return zero, nil, fmt.Errorf("failed to deserialize subject: %w", err)
	}

	return result, claims, nil
}

func VerifyWithEncryptSubjectAndClaims[T any](tokenString string, secretKey []byte) (T, jwt.Claims, error) {
	claims := &jwt.RegisteredClaims{}
	return verifyEncryptSubjectAndClaims[T](tokenString, claims, secretKey)
}

func VerifyWithEncryptSubject[T any](tokenString string, secretKey []byte) (T, error) {
	claims := &jwt.RegisteredClaims{}
	subject, _, err := verifyEncryptSubjectAndClaims[T](tokenString, claims, secretKey)
	return subject, err
}

func GenerateWithSubject(subject any, claims *jwt.RegisteredClaims, secretKey []byte) (string, error) {
	ss := e2json.MustToJSONString(subject)
	if ss == "" {
		return "", errors.New("subject is empty")
	}
	return generate(ss, claims, secretKey)
}

func generate(subject string, claims *jwt.RegisteredClaims, secretKey []byte) (string, error) {
	claims.Subject = subject
	claims.IssuedAt = jwt.NewNumericDate(time.Now())
	claims.NotBefore = jwt.NewNumericDate(time.Now())
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func VerifyWithSubject[T any](tokenString string, secretKey []byte) (T, error) {
	var zero T
	if tokenString == "" {
		return zero, errors.New("token is empty")
	}
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return zero, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return zero, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return zero, errors.New("invalid claims type")
	}
	if claims.Subject == "" {
		return zero, errors.New("subject is empty")
	}
	var result T
	if err = e2json.MustFromJSONString(claims.Subject, &result); err != nil {
		return zero, fmt.Errorf("failed to deserialize subject: %w", err)
	}
	return result, nil
}
