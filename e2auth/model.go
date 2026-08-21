package e2auth

import (
	"strings"
	"time"

	"github.com/e2u/e2util/e2db"
)

type SessionSubject struct {
	SessionId string `json:"session_id"`
	UserId    string `json:"user_id"`
}

type User struct {
	e2db.Model
	Id              string    `gorm:"column:id;type:uuid;unique;index" json:"id"`
	Name            string    `gorm:"column:name;unique;index" json:"name"`
	Email           string    `gorm:"column:email;index;unique;not null" json:"email"`
	EmailVerified   bool      `gorm:"column:email_verified;default:false" json:"email_verified"`
	PasswordHash    []byte    `gorm:"column:password_hash" json:"-"`
	Roles           []string  `gorm:"column:roles;index;type:text[]" json:"roles"`
	OTPEnable       bool      `gorm:"column:otp_enable;default:false" json:"otp_enable"`
	OTPSecret       string    `gorm:"column:otp_secret;" json:"otp_secret"`
	LastLogin       time.Time `gorm:"column:last_login" json:"last_login"`
	OAuth2Providers []string  `gorm:"column:oauth2_providers;index;type:text[]" json:"oauth2_providers"`
	ExternalID      string    `gorm:"column:external_id;index;unique;not null" json:"external_id"`
}

func (u *User) Sanitize() {
	u.Name = strings.TrimSpace(u.Name)
	u.Email = strings.TrimSpace(u.Email)
	u.ExternalID = strings.TrimSpace(u.ExternalID)
}

func (u *User) TableName() string {
	return "users"
}

type PasswordResetToken struct {
	e2db.Model
	UserId    string    `gorm:"column:user_id;type:uuid;unique;index" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserId" json:"user"`
	Token     string    `gorm:"column:token;type:text;not null;unique;index" json:"token"`
	ExpiresAt time.Time `gorm:"column:expires_at;type:timestamptz" json:"expires_at"`
}

func (pr *PasswordResetToken) Sanitize() {
	pr.Token = strings.TrimSpace(pr.Token)
	pr.UserId = strings.TrimSpace(pr.UserId)
}

func (pr *PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

type OTPRecoveryCode struct {
	e2db.Model
	UserId string `gorm:"column:user_id;type:uuid;unique;index" json:"user_id"`
	User   *User  `gorm:"foreignKey:UserId" json:"user"`
	Code   string `gorm:"column:code;not null;unique" json:"code"`
	Used   bool   `gorm:"column:used;not null;default:false" json:"used"`
}

func (or *OTPRecoveryCode) TableName() string {
	return "otp_recovery_codes"
}

type Session struct {
	e2db.Model
	SessionId string    `gorm:"column:session_id;type:uuid;unique;index" json:"session_id"`
	UserId    string    `gorm:"column:user_id;type:uuid;unique;index" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserId" json:"user"`
	Token     string    `gorm:"column:token;not null;unique;index" json:"token"`
	ExpiresAt time.Time `gorm:"column:expires_at;type:timestamptz" json:"expires_at"`
	Revoked   bool      `gorm:"column:revoked;default:false" json:"revoked"`
	IPAddress string    `gorm:"column:ip_address;" json:"ip_address"`
}

func (sr *Session) TableName() string {
	return "sessions"
}

type OAuth2Token struct {
	UserId       string    `gorm:"column:user_id;type:uuid;unique;index" json:"user_id"`
	User         *User     `gorm:"foreignKey:UserId" json:"user"`
	Provider     string    `gorm:"column:provider;" json:"provider"`
	AccessToken  string    `gorm:"column:access_token;not null" json:"access_token"`
	RefreshToken string    `gorm:"column:refresh_token;" json:"refresh_token"`
	ExpiresAt    time.Time `gorm:"column:expires_at;type:timestamptz" json:"expires_at"`
}

func (otr *OAuth2Token) TableName() string {
	return "oauth2_tokens"
}

type Configuration struct {
	e2db.Model
	Key     string `gorm:"column:key;type:text;not null;index;unique" json:"key"`
	Value   string `gorm:"column:value;type:text;not null" json:"value"`
	Comment string `gorm:"column:comment;type:text;not null" json:"comment"`
}

func (cr *Configuration) TableName() string {
	return "configurations"
}
