package e2auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/e2u/e2util/e2crypto"
	"github.com/e2u/e2util/e2exec"
	"github.com/gin-gonic/gin"
)

func verifyEmailSent(c *gin.Context) {
	user, err := getCtxUserOrAbort(c)
	if err != nil || user == nil {
		return
	}

	code := fmt.Sprintf("%06d", e2exec.Must(e2crypto.RandomNumber(100000, 999999)))
	if err := saveEmailVerifyCode(user.Id, code, time.Now().Add(durationEmailVerify)); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	data := gin.H{
		"code": code,
	}

	if err := cfg.emailer.SendTemplateEmail(user.Email, data); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}

func verifyEmailConfirm(c *gin.Context) {
	var input struct {
		Code string `json:"code" binding:"required"`
	}
	if !bindInput(c, &input) {
		return
	}

	user, err := getCtxUserOrAbort(c)
	if err != nil || user == nil {
		return
	}

	if user.EmailVerifyCode == "" || time.Now().After(user.EmailVerifyExpiresAt) || hashOpaque(input.Code) != user.EmailVerifyCode {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "Invalid or expired code"))
		return
	}

	if err := cfg.db.Model(&User{}).Where("id = ?", user.Id).Updates(map[string]any{
		"email_verified":          true,
		"email_verify_code":       "",
		"email_verify_expires_at": time.Time{},
	}).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}
