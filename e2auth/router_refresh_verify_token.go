package e2auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/e2u/e2util/e2jwt"
	"github.com/gin-gonic/gin"
)

func refreshToken(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")

	if tokenString == "" {
		c.AbortWithStatusJSON(http.StatusForbidden, errResp(ErrCodeForbidden, "Session token is empty"))
		return
	}

	subject, claims, err := e2jwt.VerifyWithEncryptSubjectAndClaims[*SessionSubject](tokenString, getSecretKey())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, errResp(ErrCodeInvalidToken, err))
		return
	}
	expire, err := claims.GetExpirationTime()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	if time.Until(expire.Time) >= 15*time.Minute {
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		c.JSON(http.StatusOK, successResp(gin.H{"token": tokenString}))
		return
	}

	duration := durationSessionLong
	token, err := generateSessionToken(subject.SessionId, subject.UserId, duration, cfg)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}

	expiresAt := time.Now().Add(duration)
	if err = updateSession(subject.SessionId, subject.UserId, token, expiresAt); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}

	data := gin.H{"token": token, "expires_at": expiresAt}
	c.JSON(http.StatusOK, successResp(data))
}

func verifyToken(c *gin.Context) {
	// Placeholder: Implement token verification logic using cfg.sessoiner
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Verify token not implemented"})
}
