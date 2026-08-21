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
		Username string `json:"username" form:"username" binding:"omitempty,min=3"`
		Password string `json:"password" form:"password" binding:"omitempty,min=8"`
		Email    string `json:"email" form:"email" binding:"omitempty,email"`
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

	if user, err := retrieveUser(cfg, input.Email, input.Username); err == nil && user != nil && (strTrimEqual(user.Email, input.Email) || strTrimEqual(user.Name, input.Username)) {
		c.AbortWithStatusJSON(http.StatusConflict, errResp(ErrCodeUserExists, errors.New("user already exists")))
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		cfg.logger.Warnf("Failed to hash password: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}

	id := uuid.NewV4().String()
	newUser := &User{
		Id:            id,
		Name:          input.Username,
		Email:         input.Email,
		EmailVerified: false,
		PasswordHash:  hashedPassword,
		ExternalID:    "local:" + id,
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
	if redirectHTML(c, "/auth/login?notice=registered") {
		return
	}
	c.JSON(http.StatusOK, successResp(data))
}
