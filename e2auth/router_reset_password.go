package e2auth

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func resetPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	input.Email = strings.TrimSpace(input.Email)

	if !bindInput(c, &input) {
		return
	}

	user, err := retrieveUser(cfg, input.Email, "")
	if user == nil || err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidCredentials, err))
		return
	}

	token, err := generateRecoverToken(uuid.NewString(), user.Id, durationRecover, cfg)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
	}

	resetPasswordToken := &PasswordResetToken{
		UserId:    user.Id,
		User:      user,
		Token:     token,
		ExpiresAt: time.Now().Add(durationRecover),
	}

	if err = createResetPasswordToken(resetPasswordToken); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}

	data := map[string]any{
		"duration_minute": durationRecover.Minutes,
		"token":           token,
		"user":            gin.H{"id": user.Id, "name": user.Name, "email": user.Email},
	}

	if err = cfg.emailer.SendTemplateEmail(user.Email, data); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(data))
}

func resetPasswordConfirm(c *gin.Context) {
	var input struct {
		Token    string `json:"token" binding:"required"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if !bindInput(c, &input) {
		return
	}

	input.Password = strings.TrimSpace(input.Password)

	if !isValidPassword(input.Password) {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidPassword, msgPasswordRule))
		return
	}

	passwordToken, err := getResetPasswordToken(input.Token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	if passwordToken.Token != input.Token {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errResp(ErrCodeInvalidToken, "Invalid token"))
		return
	}

	user, err := getUserOrAbort(c, passwordToken.UserId)
	if err != nil || user == nil {
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		cfg.logger.Warn("Failed to hash password: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}

	if err = updateUserPassword(user.Id, hashedPassword); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}
