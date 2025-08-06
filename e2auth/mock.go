package e2auth

import (
	"time"

	"github.com/e2u/e2util/e2exec"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

var mockUser = &User{
	Id:           "id-123456",
	Name:         "Tom",
	Email:        "admin@byd.io",
	PasswordHash: e2exec.Must(bcrypt.GenerateFromPassword([]byte("abcABC123!@#"), bcrypt.DefaultCost)),
}

var mockSession = &Session{
	SessionId: "token-id-123456",
	UserId:    mockUser.Id,
	User:      mockUser,
	Token:     "token-1234567890",
	ExpiresAt: time.Now().Add(time.Hour),
	Revoked:   false,
	IPAddress: "",
}

type ModeEmailer struct{}

type MockStorage struct{}

func (s *MockStorage) CreateUser(user *User) error {
	logrus.WithField("type", "MOCK").Warnf("CreateUser called")
	return nil
}
func (s *MockStorage) GetUser(id string) (*User, error) {
	logrus.WithField("type", "MOCK").Warnf("GetUser called")
	return mockUser, nil
}

func (s *MockStorage) UpdateUserPassword(id string, newPassword []byte) error {
	logrus.WithField("type", "MOCK").Warnf("UpdateUserPassword called")
	return nil
}

func (s *MockStorage) UpdateUserEmail(id string, newEmail string) error {
	logrus.WithField("type", "MOCK").Warnf("UpdateUserEmail called")
	return nil
}

func (s *MockStorage) DeleteUser(id string) error {
	logrus.WithField("type", "MOCK").Warnf("DeleteUser called")
	return nil
}
func (s *MockStorage) GetUserByEmail(email string) (*User, error) {
	logrus.WithField("type", "MOCK").Warnf("GetUserByEmail called")
	return mockUser, nil
}
func (s *MockStorage) GetUserByName(name string) (*User, error) {
	logrus.WithField("type", "MOCK").Warnf("GetUserByName called")
	return mockUser, nil
}
func (s *MockStorage) IsAdmin(id string) (bool, error) {
	logrus.WithField("type", "MOCK").Warnf("IsAdmin called")
	return false, nil
}
func (s *MockStorage) IncrementLoginFailures(id string) (int, error) {
	logrus.WithField("type", "MOCK").Warnf("IncrementLoginFailures called")
	return 0, nil
}
func (s *MockStorage) LockAccount(id string) error {
	logrus.WithField("type", "MOCK").Warnf("LockAccount called")
	return nil
}
func (s *MockStorage) UnlockAccount(id string) error {
	logrus.WithField("type", "MOCK").Warnf("UnlockAccount called")
	return nil
}
func (s *MockStorage) CreateResetPasswordToken(rp *PasswordResetToken) error {
	logrus.WithField("type", "MOCK").Warnf("CreateResetPasswordToken called")
	return nil
}

func (s *MockStorage) GetResetPasswordToken(tokenId string) (*PasswordResetToken, error) {
	logrus.WithField("type", "MOCK").Warnf("GetResetPasswordToken called")
	return &PasswordResetToken{
		UserId:    mockUser.Id,
		User:      mockUser,
		Token:     "reset-password-token",
		ExpiresAt: time.Time(time.Now().Add(durationRecover)),
	}, nil
}

func (s *MockStorage) GetSecretKey() []byte {
	logrus.WithField("type", "MOCK").Warnf("GetSecretKey called")
	return []byte("secret key")
}

func (s *MockStorage) CreateSession(session *Session) error {
	logrus.WithField("type", "MOCK").Warnf("createSession called")
	return nil
}
func (s *MockStorage) GetSession(tokenId string) (*Session, error) {
	logrus.WithField("type", "MOCK").Warnf("GetSession called")
	return mockSession, nil
}
func (s *MockStorage) RevokeSession(tokenId string) error {
	logrus.WithField("type", "MOCK").Warnf("RevokeSession called")
	return nil
}
func (s *MockStorage) RefreshSession(session *Session) error {
	logrus.WithField("type", "MOCK").Warnf("RefreshSession called")
	return nil
}
func (s *MockStorage) UpdateSession(sessionId string, userId string, token string, expiresAt time.Time) error {
	logrus.WithField("type", "MOCK").Warnf("UpdateSession called")
	return nil
}
func (s *MockStorage) UpdateEmailVerify(email string, verify bool) error {
	logrus.WithField("type", "MOCK").Warnf("ConfirmEmail called")
	return nil
}
func (s *MockStorage) SaveEmailVerifyCode(email string, code string) error {
	logrus.WithField("type", "MOCK").Warnf("SaveEmailVerifyCode called")
	return nil
}
func (s *MockStorage) FindEmailVerifyCode(email string, code string) error {
	logrus.WithField("type", "MOCK").Warnf("SaveEmailVerifyCode called")
	return nil
}
