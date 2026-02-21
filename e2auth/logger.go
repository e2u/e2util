package e2auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Logger defines the logging interface for middleware.
type Logger interface {
	Info(format string, args ...any)
	Error(format string, args ...any)
	Warn(format string, args ...any)
}

// ConsoleLogger is a simple logger that writes to stdout.
type ConsoleLogger struct{}

func (cl *ConsoleLogger) Info(format string, args ...any) {
	fmt.Printf("[INFO] %s %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(format, args...))
}
func (cl *ConsoleLogger) Error(format string, args ...any) {
	fmt.Printf("[ERROR] %s %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(format, args...))
}
func (cl *ConsoleLogger) Warn(format string, args ...any) {
	fmt.Printf("[WARN] %s %s\n", time.Now().Format(time.RFC3339), fmt.Sprintf(format, args...))
}

// NoopLogger is a no-op logger for when logging is disabled.
type NoopLogger struct{}

// Info does nothing.
func (nl *NoopLogger) Info(format string, args ...any) {}

// Error does nothing.
func (nl *NoopLogger) Error(format string, args ...any) {}

// Warn does nothing.
func (nl *NoopLogger) Warn(format string, args ...any) {}

// loggingMiddleware logs request details using a Logger interface.
func loggingMiddleware(logger Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time for latency calculation
		start := time.Now()

		// Capture request details
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()

		// Get userID if authenticated (set by authMiddleware via CertificateStorage)
		userID := "anonymous"
		if id, exists := c.Get(ctxKeyUserId); exists {
			if strID, ok := id.(string); ok {
				userID = strID
			} else {
				logger.Warn("Invalid userID type in context")
			}
		}

		// Process the request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Log request details
		status := c.Writer.Status()
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			// Set the request ID in response header for traceability
			c.Header("X-Request-ID", requestID)
		}
		logger.Info("Request: requestID=%s method=%s path=%s status=%d clientIP=%s userID=%s latency=%v",
			requestID, method, path, status, clientIP, userID, latency)

		// Log any errors
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				logger.Error("Request error: method=%s path=%s error=%s userID=%s",
					method, path, err.Error(), userID)
			}
		}

		if status == http.StatusForbidden && c.Request.Method == http.MethodPost {
			logger.Warn("CSRF validation failed: method=%s path=%s clientIP=%s userID=%s",
				method, path, clientIP, userID)
		}
	}
}
