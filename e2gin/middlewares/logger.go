package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// RequestLoggingMiddleware logs request and response details.
// Note: This reads the request body, so it should be used only for debugging
// or with appropriate performance considerations.
func RequestLoggingMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Capture request body
		var reqBody []byte
		if c.Request.Body != nil && c.Request.Method != http.MethodGet {
			var err error
			reqBody, err = io.ReadAll(c.Request.Body)
			if err == nil {
				// Restore body so handlers can read it
				c.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
			}
		}

		// Wrap response writer to capture response
		blw := &responseWriter{
			body:           &bytes.Buffer{},
			ResponseWriter: c.Writer,
		}
		c.Writer = blw

		c.Next()

		// Log after request is processed
		logger.WithFields(logrus.Fields{
			"status":       c.Writer.Status(),
			"method":       c.Request.Method,
			"path":         c.Request.URL.Path,
			"query_params": c.Request.URL.Query(),
			"req_body":     string(reqBody),
			"res_body":     blw.body.String(),
			"client_ip":    c.ClientIP(),
		}).Info("request details")
	}
}

