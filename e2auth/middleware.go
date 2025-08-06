package e2auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.AbortWithStatusJSON(http.StatusGone, errResp(ErrCodeUnauthorized, "Unauthorized"))
			return
		}

		subject, err := getSessionSubjectOrAbort(c, getSecretKey())
		if err != nil {
			c.Abort()
			return
		}

		c.Set(ctxKeyUserId, subject.UserId)
		c.Next()
	}
}

func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get(ctxKeyUserId)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errResp(ErrCodeUnauthorized, "Unauthorized"))
			return
		}
		isAdmin, err := isAmin(userID.(string))
		if err != nil || !isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, errResp(ErrCodeForbidden, "Forbidden"))
			return
		}
		c.Next()
	}
}

func rateLimitMiddleware(rateLimiter RateLimiter, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.ClientIP()
		if userID, exists := c.Get(ctxKeyUserId); exists {
			clientID = userID.(string)
		}
		allowed, err := rateLimiter.Allow(clientID, limit, window)
		if err != nil || !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, errResp(ErrCodeRateLimitExceeded, "Too many requests, try again later"))
			return
		}
		c.Next()
	}
}

// func csrfMiddleware(storage Storager) gin.HandlerFunc {
//	return func(c *gin.Context) {
//		if c.Request.Method != http.MethodPost {
//			c.Next()
//			return
//		}
//
//		csrfSubject, err := getCSRFSubjectOrAbout(c, storage.GetSecretKey())
//		if err != nil {
//			return
//		}
//
//		sessionSubject, err := getSessionSubjectOrAbort(c, storage.GetSecretKey())
//		if err != nil {
//			return
//		}
//
//		if csrfSubject.SessionId != sessionSubject.SessionId {
//			c.AbortWithStatusJSON(http.StatusForbidden, errResp(ErrCodeInvalidCSRFToken, "CSRF token is invalid"))
//			return
//		}
//
//		c.Next()
//	}
//}
