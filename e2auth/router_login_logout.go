package e2auth

import (
	"net/http"
	"strings"
	"time"
	"uuid"

	"github.com/e2u/e2util/e2jwt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func login(c *gin.Context) {
	var input struct {
		Username string `json:"username" form:"username" binding:"omitempty,min=3"`
		Password string `json:"password" form:"password" binding:"required,min=8"`
		Email    string `json:"email" form:"email" binding:"omitempty,email"`
		MFAToken string `json:"mfa_token" form:"mfa_token"`
	}

	if !bindInput(c, &input) {
		return
	}

	input.Password = strings.TrimSpace(input.Password)
	input.Email = strings.TrimSpace(input.Email)
	input.Username = strings.TrimSpace(input.Username)
	input.MFAToken = strings.TrimSpace(input.MFAToken)

	if !isValidPassword(input.Password) {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidPassword, msgPasswordRule))
		return
	}

	if input.Email == "" && input.Username == "" {
		cfg.logger.Warnf("Invalid email or username, email=%v, username=%v", input.Email, input.Username)
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidCredentials, "invalid email or username"))
		return
	}

	user, err := retrieveUser(cfg, input.Email, input.Username)
	if user == nil || err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidCredentials, err))
		return
	}

	if accountLocked(user) {
		c.AbortWithStatusJSON(http.StatusForbidden, errResp(ErrCodeAccountLocked, "account is locked"))
		return
	}

	// Verify password
	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(input.Password)); err != nil {
		cfg.logger.Warnf("Invalid password, email=%v, username=%v", input.Email, input.Username)
		_ = recordLoginFailure(user)
		c.AbortWithStatusJSON(http.StatusUnauthorized, errResp(ErrCodeInvalidCredentials, "invalid password"))
		return
	}

	if user.OTPEnable {
		if input.MFAToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errResp(ErrCodeMFARequired, "MFA token required"))
			return
		}
		if !verifyUserMFA(user, input.MFAToken) {
			_ = recordLoginFailure(user)
			c.AbortWithStatusJSON(http.StatusUnauthorized, errResp(ErrCodeInvalidCredentials, "Invalid MFA token"))
			return
		}
	}

	if err = clearLoginFailures(user.Id); err != nil {
		cfg.logger.Warnf("failed to clear login failures: %v", err)
	}

	sessionId := uuid.NewV4().String()
	// Create session
	// duration := e2var.IfElse(input.RememberMe, true, durationSession, durationSessionLong)
	duration := durationSessionLong
	token, err := generateSessionToken(sessionId, user.Id, duration, cfg)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	expiresAt := time.Now().Add(duration)

	session := &Session{
		SessionId: sessionId,
		UserId:    user.Id,
		Token:     token,
		ExpiresAt: expiresAt,
		IPAddress: c.ClientIP(),
	}
	err = createSession(session)
	if err != nil {
		cfg.logger.Errorf("Failed to create session, email=%v, username=%v: %v", input.Email, input.Username, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeSessionFailed, err))
		return
	}

	_ = cfg.eventNotifier.Notify(user.Id, "login", "User logged in")
	writeAuthSuccess(c, token, expiresAt, user)
}

func logout(c *gin.Context) {
	if token := requestSessionToken(c); token != "" {
		if subject, err := e2jwt.VerifyWithEncryptSubject[*SessionSubject](token, getSecretKey()); err == nil && subject != nil {
			_ = revokeSession(subject.SessionId)
		}
	}
	clearSessionCookie(c)
	if wantsHTML(c) {
		c.Redirect(http.StatusSeeOther, "/auth/login")
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}
