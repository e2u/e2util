package e2auth

import (
	"errors"
	"net/http"
	"strings"
	"uuid"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func register(c *gin.Context) {
	var input struct {
		Username string `json:"username" binding:"omitempty,min=3"`
		Password string `json:"password" binding:"omitempty,min=8"`
		Email    string `json:"email" binding:"omitempty,email"`
	}

	input.Password = strings.TrimSpace(input.Password)
	input.Email = strings.TrimSpace(input.Email)
	input.Username = strings.TrimSpace(input.Username)

	if !bindInput(c, &input) {
		return
	}

	if !isValidPassword(input.Password) {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidPassword, msgPasswordRule))
		return
	}

	if user, err := retrieveUser(cfg, input.Email, input.Username); err == nil && user != nil && (strTrimEqual(user.Email, input.Email) || strTrimEqual(user.Name, input.Username)) {
		c.AbortWithStatusJSON(http.StatusConflict, errResp(ErrCodeUserExists, errors.New("user already exists")))
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		cfg.logger.Warnf("Failed to hash password: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
	}

	newUser := &User{
		Id:            uuid.NewV4().String(),
		Name:          input.Username,
		Email:         input.Email,
		EmailVerified: false,
		PasswordHash:  hashedPassword,
	}
	err = createUser(newUser)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	data := gin.H{
		"user": gin.H{
			"id":    newUser.Id,
			"name":  newUser.Name,
			"email": newUser.Email,
		},
	}
	c.JSON(http.StatusOK, successResp(data))
}
