package e2auth

import (
	"time"
)

type ErrCode string

const (
	ErrCodeInvalidInput       ErrCode = "invalid_input"
	ErrCodeInvalidCredentials ErrCode = "invalid_credentials"
	ErrCodeInvalidPassword    ErrCode = "invalid_password"
	ErrCodeUserNotFound       ErrCode = "user_not_found"
	ErrCodeSessionFailed      ErrCode = "session_creation_failed"
	// ErrCodeCSRFTokenFailed     ErrCode = "csrf_token_failed"
	ErrCodeRateLimitExceeded   ErrCode = "rate_limit_exceeded"
	ErrCodeUserExists          ErrCode = "user_exists"
	ErrCodeInternalServerError ErrCode = "internal_server_error"
	ErrCodeUnauthorized        ErrCode = "unauthorized"
	// ErrCodeInvalidCSRFToken    ErrCode = "invalid_csrf_token"
	ErrCodeInvalidToken ErrCode = "invalid_token"
	ErrCodeForbidden    ErrCode = "forbidden"
)

type StatusCode string

const (
	StatusCodeError   StatusCode = "error"
	StatusCodeSuccess StatusCode = "success"
)

const (
	tokenLen = 32
)

const (
	// issuerCSRF    = "csrf"
	issuerSession = "session"
	issuerRecover = "recover"
)

const (
	durationSession     = time.Hour * 24 * 30
	durationSessionLong = time.Hour * 24 * 365
	durationRecover     = time.Minute * 15
	// durationCSRF        = time.Hour * 24
)

const (
	msgPasswordRule = "Password must be at least 8 characters with 1 uppercase, 1 lowercase, 1 digit, and 1 special character"
)

const (
	ctxKeyUserId = "user_id"
)
