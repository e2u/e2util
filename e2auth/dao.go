package e2auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"slices"
	"strings"
	"time"
)

func hashOpaque(v string) string {
	sum := sha256.Sum256([]byte(v))
	return hex.EncodeToString(sum[:])
}

func createSession(session *Session) error {
	return cfg.db.Create(session).Error
}

func getSessionByID(sessionId string) (*Session, error) {
	var session Session
	err := cfg.db.Model(&Session{}).Where("session_id = ?", sessionId).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func revokeSession(sessionId string) error {
	return cfg.db.Model(&Session{}).Where("session_id = ?", sessionId).Update("revoked", true).Error
}

func revokeSessionForUser(sessionId, userId string) error {
	return cfg.db.Model(&Session{}).Where("session_id = ? AND user_id = ?", sessionId, userId).Update("revoked", true).Error
}

func updateSession(sessionId string, userId string, token string, expiresAt time.Time) error {
	m := map[string]any{
		"token":      token,
		"expires_at": expiresAt,
	}
	return cfg.db.Model(&Session{}).Where("session_id = ? and user_id = ? ", sessionId, userId).UpdateColumns(m).Error
}

func retrieveUser(cfg *routerConfig, email, username string) (*User, error) {
	if strTrimEqual(email, "") && strTrimEqual(username, "") {
		return nil, errors.New("username and email is empty")
	}

	var err error
	var user *User

	if email != "" && strings.Contains(email, "@") {
		user, err = getUserByEmail(email)
	} else if username != "" {
		user, err = getUserByName(username)
	}
	if user == nil || err != nil {
		cfg.logger.Errorf("Failed to retrieve user, email=%v, username=%v, error=%v", email, username, err)
		return nil, err
	}
	return user, nil
}

func getUserByEmail(email string) (*User, error) {
	var user *User
	err := cfg.db.Model(&User{}).Where("email = ?", email).First(&user).Error
	return user, err
}

func getUserByName(name string) (*User, error) {
	var user *User
	err := cfg.db.Model(&User{}).Where("name = ?", name).First(&user).Error
	return user, err
}

func getUserByID(id string) (*User, error) {
	var user *User
	err := cfg.db.Model(&User{}).Where("id = ?", id).First(&user).Error
	return user, err
}

func createUser(user *User) error {
	return cfg.db.Create(user).Error
}

func createResetPasswordToken(rt *PasswordResetToken) error {
	return cfg.db.Create(rt).Error
}

func getResetPasswordToken(token string) (*PasswordResetToken, error) {
	var rt *PasswordResetToken
	err := cfg.db.Model(&PasswordResetToken{}).Where("token = ?", token).First(&rt).Error
	return rt, err
}

func updateUserPassword(userId string, hashedPassword []byte) error {
	return cfg.db.Model(&User{}).Where("id = ?", userId).Update("password_hash", hashedPassword).Error
}

func updateUserEmail(userId string, email string) error {
	return cfg.db.Model(&User{}).Where("id = ?", userId).Update("email", email).Error
}

func getSecretKey() []byte {
	return cfg.secretKey
}
func isAdmin(userId string) (bool, error) {
	var user User
	err := cfg.db.Model(&User{}).Select("roles").Where("id = ?", userId).First(&user).Error
	if err != nil {
		return false, err
	}
	if slices.Contains(user.Roles, "admin") {
		return true, nil
	}
	return false, nil
}

func accountLocked(user *User) bool {
	return user.LockedUntil != nil && user.LockedUntil.After(time.Now())
}

func recordLoginFailure(user *User) error {
	attempts := user.FailedLoginAttempts + 1
	updates := map[string]any{"failed_login_attempts": attempts}
	if attempts >= maxLoginFailures {
		until := time.Now().Add(durationLock)
		updates["locked_until"] = until
	}
	return cfg.db.Model(&User{}).Where("id = ?", user.Id).Updates(updates).Error
}

func clearLoginFailures(userId string) error {
	return cfg.db.Model(&User{}).Where("id = ?", userId).Updates(map[string]any{
		"failed_login_attempts": 0,
		"locked_until":          nil,
		"last_login":            time.Now(),
	}).Error
}

func unlockUser(userId string) error {
	return cfg.db.Model(&User{}).Where("id = ?", userId).Updates(map[string]any{
		"failed_login_attempts": 0,
		"locked_until":          nil,
	}).Error
}

func saveEmailVerifyCode(userId, code string, exp time.Time) error {
	return cfg.db.Model(&User{}).Where("id = ?", userId).Updates(map[string]any{
		"email_verify_code":       hashOpaque(code),
		"email_verify_expires_at": exp,
	}).Error
}

func replaceRecoveryCodes(userId string, plaintext []string) error {
	if err := cfg.db.Where("user_id = ?", userId).Delete(&OTPRecoveryCode{}).Error; err != nil {
		return err
	}
	rows := make([]OTPRecoveryCode, 0, len(plaintext))
	for _, code := range plaintext {
		rows = append(rows, OTPRecoveryCode{
			UserId: userId,
			Code:   hashOpaque(code),
			Used:   false,
		})
	}
	if len(rows) == 0 {
		return nil
	}
	return cfg.db.Create(&rows).Error
}

func consumeRecoveryCode(userId, code string) (bool, error) {
	hashed := hashOpaque(code)
	tx := cfg.db.Model(&OTPRecoveryCode{}).
		Where("user_id = ? AND code = ? AND used = ?", userId, hashed, false).
		Update("used", true)
	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, nil
}

func upsertOAuthToken(tok *OAuth2Token) error {
	var existing OAuth2Token
	err := cfg.db.Where("user_id = ? AND provider = ?", tok.UserId, tok.Provider).First(&existing).Error
	if err != nil {
		return cfg.db.Create(tok).Error
	}
	return cfg.db.Model(&existing).Updates(map[string]any{
		"access_token":  tok.AccessToken,
		"refresh_token": tok.RefreshToken,
		"expires_at":    tok.ExpiresAt,
	}).Error
}

func deleteOAuthToken(userId, provider string) error {
	return cfg.db.Where("user_id = ? AND provider = ?", userId, provider).Delete(&OAuth2Token{}).Error
}

func deleteResetTokensForUser(userId string) error {
	return cfg.db.Where("user_id = ?", userId).Delete(&PasswordResetToken{}).Error
}
