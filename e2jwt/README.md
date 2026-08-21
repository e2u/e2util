# e2jwt

用 HMAC 簽發／驗證 JWT；可選擇把 subject JSON 用 AES-GCM 加密後放進 `sub`。

HMAC JWT issue/verify, with optional AES-GCM encryption of the JSON subject in `sub`.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2jwt
```

## 功能 / Features

- **明文 subject / Plain subject**：`GenerateWithSubject`、`VerifyWithSubject`
- **加密 subject / Encrypted subject**：`GenerateWithEncryptSubject`、`VerifyWithEncryptSubject`、`VerifyWithEncryptSubjectAndClaims`

## 用法 / Usage

```go
import (
    "time"
    "github.com/e2u/e2util/e2jwt"
    "github.com/golang-jwt/jwt/v5"
)

type Sub struct {
    UserID string `json:"user_id"`
}

secret := []byte("secret-key")
token, err := e2jwt.GenerateWithEncryptSubject(Sub{UserID: "u1"}, &jwt.RegisteredClaims{
    ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
}, secret)

sub, err := e2jwt.VerifyWithEncryptSubject[Sub](token, secret)
```
