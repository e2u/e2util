package e2auth

import (
	"errors"
	"net/http"
	"net/url"
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
	var err error
	if strings.Contains(c.ContentType(), "json") {
		err = c.ShouldBindJSON(t)
	} else {
		err = c.ShouldBind(t)
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, err))
		return false
	}
	return true
}

func requestSessionToken(c *gin.Context) string {
	token := strings.TrimSpace(c.GetHeader("Authorization"))
	token = strings.TrimPrefix(token, "Bearer ")
	token = strings.TrimSpace(token)
	if token != "" {
		return token
	}
	cookie, _ := c.Cookie(sessionCookie)
	return strings.TrimSpace(cookie)
}

func wantsHTML(c *gin.Context) bool {
	return strings.Contains(c.GetHeader("Accept"), "text/html")
}

func isJSONRequest(c *gin.Context) bool {
	return strings.Contains(c.ContentType(), "json")
}

func safeNext(vals ...string) string {
	for _, v := range vals {
		v = strings.TrimSpace(v)
		if strings.HasPrefix(v, "/") && !strings.HasPrefix(v, "//") {
			return v
		}
	}
	return "/"
}

func setSessionCookie(c *gin.Context, token string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 1 {
		maxAge = 1
	}
	secure := c.Request.TLS != nil
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(sessionCookie, token, maxAge, "/", "", secure, true)
}

func clearSessionCookie(c *gin.Context) {
	secure := c.Request.TLS != nil
	c.SetCookie(sessionCookie, "", -1, "/", "", secure, true)
}

func writeAuthSuccess(c *gin.Context, token string, expiresAt time.Time, user *User) {
	setSessionCookie(c, token, expiresAt)
	data := gin.H{
		"token":      token,
		"expires_at": expiresAt,
		"user": gin.H{
			"id":    user.Id,
			"name":  user.Name,
			"email": user.Email,
		},
	}
	if wantsHTML(c) && !isJSONRequest(c) {
		c.Redirect(http.StatusSeeOther, safeNext(c.Query("next"), c.PostForm("next")))
		return
	}
	c.JSON(http.StatusOK, successResp(data))
}

func redirectHTML(c *gin.Context, location string) bool {
	if wantsHTML(c) && !isJSONRequest(c) {
		c.Redirect(http.StatusSeeOther, location)
		return true
	}
	return false
}

func abortUnauthorized(c *gin.Context, code ErrCode, message any) {
	if wantsHTML(c) && c.Request.Method == http.MethodGet {
		next := url.QueryEscape(c.Request.URL.RequestURI())
		c.Redirect(http.StatusFound, "/auth/login?next="+next)
		c.Abort()
		return
	}
	status := http.StatusUnauthorized
	if code == ErrCodeForbidden {
		status = http.StatusForbidden
	}
	c.AbortWithStatusJSON(status, errResp(code, message))
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
	sessionToken := requestSessionToken(c)
	if sessionToken == "" {
		abortUnauthorized(c, ErrCodeUnauthorized, "Session token is empty")
		return nil, errors.New("session token is empty")
	}

	subject, err := e2jwt.VerifyWithEncryptSubject[*SessionSubject](sessionToken, secretKey)
	if err != nil {
		abortUnauthorized(c, ErrCodeInvalidToken, "Invalid token")
		return nil, errors.New("invalid token")
	}
	if subject == nil {
		abortUnauthorized(c, ErrCodeInvalidToken, "Invalid token")
		return nil, errors.New("invalid token")
	}
	return subject, nil
}
