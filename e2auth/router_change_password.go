package e2auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func changePassword(c *gin.Context) {
	var input struct {
		CurrentPassword string `json:"current_password" binding:"required,min=8"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}
	if !bindInput(c, &input) {
		return
	}

	input.CurrentPassword = strings.TrimSpace(input.CurrentPassword)
	input.NewPassword = strings.TrimSpace(input.NewPassword)

	if !isValidPassword(input.NewPassword) {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidPassword, msgPasswordRule))
		return
	}

	user, err := getCtxUserOrAbort(c)
	if err != nil {
		return
	}
	// verify is correct of current password
	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(input.CurrentPassword)); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidPassword, "Invalid current password"))
		return
	}

	// new password can't equal to old password
	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(input.NewPassword)); err == nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidPassword, "New password cannot be the same as the current password"))
		return
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}

	if err = updateUserPassword(user.Id, newPasswordHash); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, "Failed to update password"))
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}
