package e2auth

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/e2u/e2util/e2jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	specialChars := "!@#$%^&*"
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case strings.ContainsRune(specialChars, char):
			hasSpecial = true
		default:
			return false
		}
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}

func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func strTrimEqual(a, b string) bool {
	return strings.EqualFold(strings.TrimSpace(a), strings.TrimSpace(b))
}

func bindInput(c *gin.Context, t any) bool {
	if err := c.ShouldBindJSON(&t); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, err))
		return false
	}
	return true
}

// func generateCSRFToken(tokenId, sessionId string, duration time.Duration, cfg *routerConfig) (string, error) {
//	subject := &CSRFSubject{
//		SessionId: sessionId,
//	}
//	claims := &jwt.RegisteredClaims{
//		ID:        tokenId,
//		Issuer:    issuerCSRF,
//		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
//		NotBefore: jwt.NewNumericDate(time.Now()),
//	}
//
//	token, err := e2jwt.GenerateWithEncryptSubject(subject, claims, cfg.storager.GetSecretKey())
//	if err != nil {
//		return "", err
//	}
//	return token, nil
//}

func generateSessionToken(sessionId, userId string, duration time.Duration, cfg *routerConfig) (string, error) {
	subject := &SessionSubject{
		SessionId: sessionId,
		UserId:    userId,
	}
	claims := &jwt.RegisteredClaims{
		ID:        sessionId,
		Issuer:    issuerSession,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
	}
	token, err := e2jwt.GenerateWithEncryptSubject(subject, claims, getSecretKey())
	if err != nil {
		return "", err
	}
	return token, nil
}

func generateRecoverToken(sessionId, userId string, duration time.Duration, cfg *routerConfig) (string, error) {
	return generateSessionToken(sessionId, userId, duration, cfg)
}

// func getVerifiedSessionSubject(c *gin.Context, secretKey []byte) (*SessionSubject, error) {
//	sessionToken := c.GetHeader("Authorization")
//	if sessionToken == "" {
//		return nil, errors.New("session token is empty")
//	}
//	subject, err := e2jwt.VerifyWithEncryptSubject[*SessionSubject](sessionToken, secretKey)
//	if err != nil || subject == nil {
//		return nil, errors.New("invalid session token")
//	}
//	return subject, nil
//}

// func getVerifiedCSRFSubject(c *gin.Context, secretKey []byte) (*CSRFSubject, error) {
//	csrfToken := c.GetHeader("X-CSRF-Token")
//	if csrfToken == "" {
//		csrfToken = c.PostForm("csrf_token")
//	}
//	if csrfToken == "" {
//		return nil, errors.New("invalid csrf token")
//	}
//	subject, err := e2jwt.VerifyWithEncryptSubject[*CSRFSubject](csrfToken, secretKey)
//	if err != nil || subject == nil {
//		return nil, errors.New("invalid csrf token")
//	}
//	return subject, nil
//}

func getSessionSubjectOrAbort(c *gin.Context, secretKey []byte) (*SessionSubject, error) {
	sessionToken := c.GetHeader("Authorization")
	if sessionToken == "" {
		c.AbortWithStatusJSON(http.StatusForbidden, errResp(ErrCodeForbidden, "Session token is empty"))
		return nil, errors.New("session token is empty")
	}

	subject, err := e2jwt.VerifyWithEncryptSubject[*SessionSubject](sessionToken, secretKey)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, errResp(ErrCodeInvalidToken, "Invalid token"))
		return nil, errors.New("invalid token")
	}
	if subject == nil {
		c.AbortWithStatusJSON(http.StatusForbidden, errResp(ErrCodeInvalidToken, "Invalid token"))
		return nil, errors.New("invalid token")
	}
	return subject, nil
}
