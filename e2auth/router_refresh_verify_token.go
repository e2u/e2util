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
	user, err := getCtxUserOrAbort(c)
	if err != nil {
		return
	}
	sessionID, _ := c.Get(ctxKeySessionId)
	data := gin.H{
		"valid": true,
		"user": gin.H{
			"id":             user.Id,
			"name":           user.Name,
			"email":          user.Email,
			"roles":          user.Roles,
			"email_verified": user.EmailVerified,
			"otp_enable":     user.OTPEnable,
		},
	}
	if sid, ok := sessionID.(string); ok && sid != "" {
		data["session_id"] = sid
	}
	c.JSON(http.StatusOK, successResp(data))
}
