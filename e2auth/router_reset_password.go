package e2auth

import (
	"net/http"
	"strings"
	"time"
	"uuid"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func resetPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email" form:"email" binding:"required,email"`
	}
	if !bindInput(c, &input) {
		return
	}
	input.Email = strings.TrimSpace(input.Email)

	user, err := retrieveUser(cfg, input.Email, "")
	if user == nil || err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidCredentials, err))
		return
	}

	token, err := generateRecoverToken(uuid.NewV4().String(), user.Id, durationRecover, cfg)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}

	_ = deleteResetTokensForUser(user.Id)
	resetPasswordToken := &PasswordResetToken{
		UserId:    user.Id,
		Token:     token,
		ExpiresAt: time.Now().Add(durationRecover),
	}

	if err = createResetPasswordToken(resetPasswordToken); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}

	data := map[string]any{
		"duration_minute": durationRecover.Minutes(),
		"token":           token,
		"reset_url":       "/auth/reset-password?token=" + token,
		"user":            gin.H{"id": user.Id, "name": user.Name, "email": user.Email},
	}

	if err = cfg.emailer.SendTemplateEmail(user.Email, data); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	if redirectHTML(c, "/auth/login?notice=reset_sent") {
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}

func resetPasswordConfirm(c *gin.Context) {
	var input struct {
		Token    string `json:"token" form:"token" binding:"required"`
		Password string `json:"password" form:"password" binding:"required,min=8"`
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
		c.AbortWithStatusJSON(http.StatusUnauthorized, errResp(ErrCodeInvalidToken, "Invalid token"))
		return
	}
	if passwordToken.Token != input.Token || time.Now().After(passwordToken.ExpiresAt) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errResp(ErrCodeInvalidToken, "Invalid or expired token"))
		return
	}

	user, err := getUserOrAbort(c, passwordToken.UserId)
	if err != nil || user == nil {
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		cfg.logger.Warnf("Failed to hash password: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}

	if err = updateUserPassword(user.Id, hashedPassword); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	_ = deleteResetTokensForUser(user.Id)
	if redirectHTML(c, "/auth/login?notice=password_updated") {
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}
