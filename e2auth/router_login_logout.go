package e2auth

import (
	"net/http"
	"strings"
	"time"
	"uuid"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func login(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"omitempty,min=3"`
		Password string `json:"password" binding:"required,min=8"`
		Email    string `json:"email" binding:"omitempty,email"`
	}

	if !bindInput(c, &input) {
		return
	}

	input.Password = strings.TrimSpace(input.Password)
	input.Email = strings.TrimSpace(input.Email)
	input.Username = strings.TrimSpace(input.Username)

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

	// Verify password
	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(input.Password)); err != nil {
		cfg.logger.Warnf("Invalid password, email=%v, username=%v", input.Email, input.Username)
		c.AbortWithStatusJSON(http.StatusUnauthorized, errResp(ErrCodeInvalidCredentials, "invalid password"))
		return
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
		User:      user,
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

	// Notify login event
	_ = cfg.eventNotifier.Notify(user.Id, "login", "User logged in")

	data := gin.H{
		"token":      token,
		"expires_at": expiresAt,
		"user": gin.H{
			"id":    user.Id,
			"name":  user.Name,
			"email": user.Email,
		},
	}
	c.JSON(http.StatusOK, successResp(data))
}

func logout(c *gin.Context) {
	subject, err := getSessionSubjectOrAbort(c, getSecretKey())
	if err != nil {
		return
	}
	if err = revokeSession(subject.SessionId); err != nil {
		c.AbortWithStatusJSON(http.StatusForbidden, errResp(ErrCodeInvalidToken, "Invalid token"))
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}
