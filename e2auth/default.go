package e2auth

import (
	"time"
)

type noopEventNotifier struct{}

func (nen *noopEventNotifier) Notify(userID, eventType, message string) error { return nil }

type noopLogger struct{}

// func (nl *noopLogger) Info(format string, args ...any)   {}
// func (nl *noopLogger) Error(format string, args ...any)  {}
// func (nl *noopLogger) Warn(format string, args ...any)   {}
func (nl *noopLogger) Infof(format string, args ...any)  {}
func (nl *noopLogger) Errorf(format string, args ...any) {}
func (nl *noopLogger) Warnf(format string, args ...any)  {}

type noopEmailer struct{}

func (ne *noopEmailer) SendEmail(to, subject, body string) error { return nil }
func (ne *noopEmailer) SendTemplateEmail(to string, data map[string]any) error {
	return nil
}

//nolint:unused
type noopOAuthProvider struct{}

//nolint:unused
func (nop *noopOAuthProvider) GetAuthURL(state string) string { return "" }

//nolint:unused
func (nop *noopOAuthProvider) ExchangeCode(code, redirectURI string) (string, string, error) {
	return "", "", nil
}

//nolint:unused
func (nop *noopOAuthProvider) GetUserInfo(accessToken string) (map[string]any, error) {
	return map[string]any{}, nil
}

type noopCAPTCHAService struct{}

func (nc *noopCAPTCHAService) Verify(response, clientIP string) (bool, error) { return true, nil }

type noopRateLimiter struct{}

func (nr *noopRateLimiter) Allow(clientID string, limit int, window time.Duration) (bool, error) {
	return true, nil
}
