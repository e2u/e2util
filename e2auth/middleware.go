package e2auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func authenticateRequest(c *gin.Context) bool {
	subject, err := getSessionSubjectOrAbort(c, getSecretKey())
	if err != nil {
		return false
	}

	session, err := getSessionByID(subject.SessionId)
	if err != nil || session == nil || session.Revoked || time.Now().After(session.ExpiresAt) {
		abortUnauthorized(c, ErrCodeUnauthorized, "Session is invalid")
		return false
	}

	c.Set(ctxKeyUserId, subject.UserId)
	c.Set(ctxKeySessionId, subject.SessionId)
	return true
}

func authorizeAdmin(c *gin.Context) bool {
	userID, exists := c.Get(ctxKeyUserId)
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errResp(ErrCodeUnauthorized, "Unauthorized"))
		return false
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, "Invalid user ID type in context"))
		return false
	}
	okAdmin, err := isAdmin(userIDStr)
	if err != nil || !okAdmin {
		c.AbortWithStatusJSON(http.StatusForbidden, errResp(ErrCodeForbidden, "Forbidden"))
		return false
	}
	return true
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !authenticateRequest(c) {
			return
		}
		c.Next()
	}
}

func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !authorizeAdmin(c) {
			return
		}
		c.Next()
	}
}

// Required protects an application route. Use after Mount/Register.
func Required() gin.HandlerFunc {
	return authMiddleware()
}

// AdminOnly requires an admin role. Place after Required, or use RequireAdmin.
func AdminOnly() gin.HandlerFunc {
	return adminMiddleware()
}

// RequireAdmin authenticates and requires the admin role in one middleware.
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !authenticateRequest(c) {
			return
		}
		if !authorizeAdmin(c) {
			return
		}
		c.Next()
	}
}

func rateLimitMiddleware(rateLimiter RateLimiter, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientID := c.ClientIP()
		// Note: This only works for userID if the middleware is placed AFTER authMiddleware.
		// For routes using this before auth, it will use ClientIP for rate limiting.
		if userID, exists := c.Get(ctxKeyUserId); exists {
			if userIDStr, ok := userID.(string); ok {
				clientID = userIDStr
			}
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
