package middlewares

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDefaultSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eng := gin.New()
	eng.Use(DefaultSecurityHeaders())
	eng.GET("/", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	eng.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "SAMEORIGIN", w.Header().Get("X-Frame-Options"))
	assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "max-age=")
	csp := w.Header().Get("Content-Security-Policy")
	assert.Contains(t, csp, "default-src 'self'")
	assert.Contains(t, csp, "script-src")
	assert.Contains(t, csp, "https://challenges.cloudflare.com")
}

func TestSecurityHeadersCustom(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eng := gin.New()
	eng.Use(SecurityHeaders(SecurityHeadersConfig{
		ScriptSrc:               ScriptStyleSrc{HostSrc: HostSrc{Self: true}, UnsafeInline: true},
		XFrameOptions:           "DENY",
		StrictTransportSecurity: "max-age=60",
		OtherHeaders:            map[string]string{"X-Custom": "1"},
	}))
	eng.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	eng.ServeHTTP(w, req)
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "max-age=60", w.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, "1", w.Header().Get("X-Custom"))
	assert.Contains(t, w.Header().Get("Content-Security-Policy"), "'unsafe-inline'")
}

func TestRequestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	eng := gin.New()
	eng.Use(RequestLoggingMiddleware(logger))
	eng.POST("/echo", func(c *gin.Context) { c.String(http.StatusOK, "pong") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/echo?q=1", strings.NewReader("ping"))
	eng.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	logs := buf.String()
	assert.Contains(t, logs, "request details")
	assert.Contains(t, logs, "ping")
	assert.Contains(t, logs, "pong")
}
