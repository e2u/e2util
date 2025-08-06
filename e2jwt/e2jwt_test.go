package e2jwt

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

type S struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func Test_WithSubject(t *testing.T) {
	secretKey := []byte("secret")
	claims := &jwt.RegisteredClaims{
		Issuer: "test",
	}
	subject := &S{
		Id:   "id-12345678",
		Name: "test-name",
	}
	t.Run("WithSubject", func(t *testing.T) {
		token, err := GenerateWithSubject(subject, claims, secretKey)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("generate token:", token)
		s, err := VerifyWithSubject[*S](token, secretKey)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("claims: %#v", s)
	})
}
