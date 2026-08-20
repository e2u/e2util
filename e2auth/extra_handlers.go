package e2auth

import (
	"net/http"
	"time"
	"uuid"

	"github.com/e2u/e2util/e2crypto"
	"github.com/gin-gonic/gin"
)

// Unlock a user account (placeholder implementation)
func unlockAccount(c *gin.Context) {
	c.JSON(http.StatusOK, successResp(nil))
}

// Initiate OAuth flow for a provider
func oauthProvider(c *gin.Context) {
	providerName := c.Param("provider")
	provider, ok := cfg.oauthProviders[providerName]
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "unknown provider"))
		return
	}
	state := uuid.NewV4().String()
	url := provider.GetAuthURL(state)
	c.JSON(http.StatusOK, successResp(gin.H{"auth_url": url, "state": state}))
}

// Handle OAuth callback
func oauthCallback(c *gin.Context) {
	providerName := c.Param("provider")
	code := c.Query("code")
	provider, ok := cfg.oauthProviders[providerName]
	if !ok || code == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "invalid callback"))
		return
	}
	accessToken, refreshToken, err := provider.ExchangeCode(code, "")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	userInfo, err := provider.GetUserInfo(accessToken)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(gin.H{"access_token": accessToken, "refresh_token": refreshToken, "user": userInfo}))
}

// Verify CAPTCHA response
func captchaVerify(c *gin.Context) {
	var input struct {
		Response string `json:"response" binding:"required"`
	}
	if !bindInput(c, &input) {
		return
	}
	ok, err := cfg.captchaService.Verify(input.Response, c.ClientIP())
	if err != nil || !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "captcha verification failed"))
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}

// Delete current user's account
func deleteAccount(c *gin.Context) {
	user, err := getCtxUserOrAbort(c)
	if err != nil {
		return
	}
	if err := cfg.db.Delete(&User{}, "id = ?", user.Id).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	_ = cfg.db.Model(&Session{}).Where("user_id = ?", user.Id).Update("revoked", true).Error
	c.JSON(http.StatusOK, successResp(nil))
}

// Enable MFA for current user
func mfaEnable(c *gin.Context) {
	user, err := getCtxUserOrAbort(c)
	if err != nil {
		return
	}
	secret, err := e2crypto.GenerateTOTPSecret()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	if err := cfg.db.Model(&User{}).Where("id = ?", user.Id).Updates(map[string]any{"otp_enable": true, "otp_secret": secret}).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(gin.H{"otp_secret": secret}))
}

// Verify MFA token using TOTP
func mfaVerify(c *gin.Context) {
	var input struct {
		Token string `json:"token" binding:"required"`
	}
	if !bindInput(c, &input) {
		return
	}
	user, err := getCtxUserOrAbort(c)
	if err != nil {
		return
	}
	if !user.OTPEnable {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "MFA not enabled"))
		return
	}
	// Verify TOTP token with 30-second time step and 1 step skew (±30 seconds)
	if !e2crypto.VerifyTOTPWithConfig(e2crypto.DefaultTOTPConfig(user.OTPSecret), input.Token, time.Now(), 30) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errResp(ErrCodeInvalidCredentials, "Invalid MFA token"))
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}

// Disable MFA for current user
func mfaDisable(c *gin.Context) {
	user, err := getCtxUserOrAbort(c)
	if err != nil {
		return
	}
	if err := cfg.db.Model(&User{}).Where("id = ?", user.Id).Updates(map[string]any{"otp_enable": false, "otp_secret": ""}).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}

// Link OAuth account (placeholder)
func linkOAuth(c *gin.Context) {
	c.JSON(http.StatusOK, successResp(nil))
}

// Unlink OAuth account (placeholder)
func unlinkOAuth(c *gin.Context) {
	c.JSON(http.StatusOK, successResp(nil))
}

// Generate MFA backup codes
func generateMFABackupCodes(c *gin.Context) {
	codes := []string{}
	for range 5 {
		code, _ := e2crypto.RandomString(8)
		codes = append(codes, code)
	}
	c.JSON(http.StatusOK, successResp(gin.H{"codes": codes}))
}
