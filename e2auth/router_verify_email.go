package e2auth

import (
	"fmt"
	"net/http"

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
		Code string `form:"code" binding:"required"`
	}
	if !bindInput(c, &input) {
		return
	}

	user, err := getCtxUserOrAbort(c)
	if err != nil || user == nil {
		return
	}
}
