package e2auth

import (
	"time"
)

type Emailer interface {
	SendEmail(to, subject, body string) error
	SendTemplateEmail(to string, data map[string]any) error
}

// type Storager interface {
//	CreateUser(user *User) error
//	GetUser(id string) (*User, error)
//	UpdateUserPassword(id string, newPassword []byte) error
//	UpdateUserEmail(id string, newEmail string) error
//	DeleteUser(id string) error
//	GetUserByEmail(email string) (*User, error)
//	GetUserByName(name string) (*User, error)
//	IsAdmin(id string) (bool, error)
//	SaveEmailVerifyCode(email string, code string) error
//	UpdateEmailVerify(email string, verify bool) error
//
//	IncrementLoginFailures(id string) (int, error)
//	LockAccount(id string) error
//	UnlockAccount(id string) error
//
//	CreateResetPasswordToken(rp *PasswordResetToken) error
//	GetResetPasswordToken(id string) (*PasswordResetToken, error)
//
//	GetSecretKey() []byte
//
//	CreateSession(session *Session) error
//	UpdateSession(sessionId string, userId string, token string, expiresAt time.Time) error
//	GetSession(tokenId string) (*Session, error)
//	RevokeSession(tokenId string) error
//	RefreshSession(session *Session) error
//}

type CAPTCHAService interface {
	Verify(response, clientIP string) (bool, error)
}

type OAuthProvider interface {
	GetAuthURL(state string) string
	ExchangeCode(code, redirectURI string) (accessToken, refreshToken string, err error)
	GetUserInfo(accessToken string) (map[string]interface{}, error)
}

type OAuthProviders map[string]OAuthProvider

type RateLimiter interface {
	Allow(clientID string, limit int, window time.Duration) (bool, error)
}

type EventNotifier interface {
	Notify(userID, eventType, message string) error
}
