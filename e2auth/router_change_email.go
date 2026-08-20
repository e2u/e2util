package e2auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"uuid"

	"github.com/e2u/e2util/e2crypto"
	"github.com/e2u/e2util/e2exec"
	"github.com/e2u/e2util/e2jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type ChangeEmailSubject struct {
	CurrentEmail string `json:"current_email"`
	NewEmail     string `json:"new_email"`
	Code         string `json:"code"`
}

func changeEmail(c *gin.Context) {
	var input struct {
		Password     string `json:"password" binding:"required,min=8"`
		CurrentEmail string `json:"current_email" binding:"required,email"`
		NewEmail     string `json:"new_email" binding:"required,email"`
	}
	if !bindInput(c, &input) {
		return
	}
	input.Password = strings.TrimSpace(input.Password)
	input.CurrentEmail = strings.TrimSpace(input.CurrentEmail)
	input.NewEmail = strings.TrimSpace(input.NewEmail)

	if !isValidEmail(input.CurrentEmail) {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "Invalid email"))
		return
	}

	if !isValidEmail(input.NewEmail) {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "Invalid email"))
		return
	}

	if strings.EqualFold(input.NewEmail, input.CurrentEmail) {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "The new email cannot be same as the current email"))
		return
	}

	user, err := getCtxUserOrAbort(c)
	if err != nil {
		return
	}

	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(input.Password)); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, err))
		return
	}

	if !strings.EqualFold(user.Email, input.CurrentEmail) {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "The current email does not match"))
		return
	}

	code := fmt.Sprintf("%06d", e2exec.Must(e2crypto.RandomNumber(100000, 999999)))
	subject := &ChangeEmailSubject{
		CurrentEmail: input.CurrentEmail,
		NewEmail:     input.NewEmail,
		Code:         code,
	}

	claims := &jwt.RegisteredClaims{
		ID:        uuid.NewV4().String(),
		Issuer:    issuerRecover,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(durationRecover)),
	}
	token, err := e2jwt.GenerateWithEncryptSubject(subject, claims, getSecretKey())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}

	mailData := gin.H{
		"code":               code,
		"change_email_token": token,
	}

	if err = cfg.emailer.SendTemplateEmail(input.NewEmail, mailData); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}

func changeEmailConfirm(c *gin.Context) {
	var input struct {
		Code             string `json:"code" binding:"required"`
		ChangeEmailToken string `json:"change_email_token" binding:"required"`
	}

	if !bindInput(c, &input) {
		return
	}

	subject, err := e2jwt.VerifyWithEncryptSubject[*ChangeEmailSubject](input.ChangeEmailToken, getSecretKey())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidToken, err))
		return
	}

	user, err := getCtxUserOrAbort(c)
	if err != nil {
		return
	}

	if strings.EqualFold(user.Email, subject.CurrentEmail) {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "The current email does not match"))
		return
	}

	if input.Code != subject.Code {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "The code does not match"))
		return
	}

	if err = updateUserEmail(user.Id, subject.CurrentEmail); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}
