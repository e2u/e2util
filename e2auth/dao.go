package e2auth

import (
	"errors"
	"strings"
	"time"
)

func createSession(session *Session) error {
	return cfg.db.Create(session).Error
}

func revokeSession(sessionId string) error {
	return cfg.db.Model(&Session{}).Delete("session_id", sessionId).Error
}

func updateSession(sessionId string, userId string, token string, expiresAt time.Time) error {
	m := map[string]interface{}{
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
		cfg.logger.Error("Failed to retrieve user, email=%v, username=%v, error=%v", email, username, err)
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
func isAmin(userId string) (bool, error) { // TODO
	return false, nil
}
