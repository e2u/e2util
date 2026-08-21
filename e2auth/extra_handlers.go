package e2auth

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"
	"uuid"

	"github.com/e2u/e2util/e2crypto"
	"github.com/e2u/e2util/e2jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UnlockSubject struct {
	UserId string `json:"user_id"`
}

func verifyUserMFA(user *User, token string) bool {
	if user.OTPSecret != "" && e2crypto.VerifyTOTPWithConfig(e2crypto.DefaultTOTPConfig(user.OTPSecret), token, time.Now(), 30) {
		return true
	}
	ok, err := consumeRecoveryCode(user.Id, token)
	return err == nil && ok
}

func unlockAccount(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
		Token string `json:"token"`
	}
	if !bindInput(c, &input) {
		return
	}
	input.Email = strings.TrimSpace(input.Email)
	input.Token = strings.TrimSpace(input.Token)

	if input.Token != "" {
		subject, err := e2jwt.VerifyWithEncryptSubject[*UnlockSubject](input.Token, getSecretKey())
		if err != nil || subject == nil || subject.UserId == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, errResp(ErrCodeInvalidToken, "Invalid unlock token"))
			return
		}
		if err := unlockUser(subject.UserId); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
			return
		}
		c.JSON(http.StatusOK, successResp(nil))
		return
	}

	user, err := getUserByEmail(input.Email)
	if err != nil || user == nil {
		// Do not leak whether the account exists.
		c.JSON(http.StatusOK, successResp(nil))
		return
	}
	if !accountLocked(user) {
		_ = unlockUser(user.Id)
		c.JSON(http.StatusOK, successResp(nil))
		return
	}

	claims := &jwt.RegisteredClaims{
		ID:        uuid.NewV4().String(),
		Issuer:    issuerUnlock,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(durationRecover)),
	}
	token, err := e2jwt.GenerateWithEncryptSubject(&UnlockSubject{UserId: user.Id}, claims, getSecretKey())
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	if err := cfg.emailer.SendTemplateEmail(user.Email, map[string]any{"token": token, "purpose": "unlock"}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}

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

func oauthString(info map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := info[k]; ok {
			switch t := v.(type) {
			case string:
				if strings.TrimSpace(t) != "" {
					return strings.TrimSpace(t)
				}
			}
		}
	}
	return ""
}

func issueSession(c *gin.Context, user *User) {
	sessionId := uuid.NewV4().String()
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
		Token:     token,
		ExpiresAt: expiresAt,
		IPAddress: c.ClientIP(),
	}
	if err = createSession(session); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeSessionFailed, err))
		return
	}
	_ = cfg.eventNotifier.Notify(user.Id, "login", "User logged in")
	writeAuthSuccess(c, token, expiresAt, user)
}

func findOrCreateOAuthUser(providerName, accessToken, refreshToken string, userInfo map[string]any) (*User, error) {
	email := oauthString(userInfo, "email", "Email")
	extID := oauthString(userInfo, "id", "sub", "user_id")
	name := oauthString(userInfo, "name", "login", "preferred_username")
	if extID == "" {
		extID = email
	}
	externalID := providerName + ":" + extID

	var existing OAuth2Token
	if err := cfg.db.Where("provider = ? AND user_id IN (SELECT id FROM users WHERE external_id = ?)", providerName, externalID).First(&existing).Error; err == nil {
		return getUserByID(existing.UserId)
	}

	var user *User
	var err error
	if email != "" {
		user, err = getUserByEmail(email)
	}
	if user == nil || err != nil {
		if name == "" {
			name = email
		}
		if name == "" {
			name = providerName + "-" + extID
		}
		id := uuid.NewV4().String()
		randPw, _ := e2crypto.RandomString(24)
		hash, herr := bcrypt.GenerateFromPassword([]byte(randPw+"Aa1!"), bcrypt.DefaultCost)
		if herr != nil {
			return nil, herr
		}
		user = &User{
			Id:              id,
			Name:            name,
			Email:           email,
			EmailVerified:   email != "",
			PasswordHash:    hash,
			ExternalID:      externalID,
			OAuth2Providers: []string{providerName},
		}
		if err = createUser(user); err != nil {
			return nil, err
		}
	} else if !slices.Contains(user.OAuth2Providers, providerName) {
		providers := append(append([]string{}, user.OAuth2Providers...), providerName)
		_ = cfg.db.Model(&User{}).Where("id = ?", user.Id).Updates(map[string]any{
			"oauth2_providers": providers,
			"external_id":      firstNonEmpty(user.ExternalID, externalID),
		}).Error
		user.OAuth2Providers = providers
	}

	if err := upsertOAuthToken(&OAuth2Token{
		UserId:       user.Id,
		Provider:     providerName,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(durationSession),
	}); err != nil {
		return nil, err
	}
	return user, nil
}

func firstNonEmpty(v ...string) string {
	for _, s := range v {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

func oauthCallback(c *gin.Context) {
	providerName := c.Param("provider")
	code := c.Query("code")
	if code == "" {
		code = c.PostForm("code")
	}
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
	user, err := findOrCreateOAuthUser(providerName, accessToken, refreshToken, userInfo)
	if err != nil || user == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	issueSession(c, user)
}

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

func mfaEnable(c *gin.Context) {
	user, err := getCtxUserOrAbort(c)
	if err != nil {
		return
	}
	if user.OTPEnable {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "MFA already enabled"))
		return
	}
	secret, err := e2crypto.GenerateTOTPSecret()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	if err := cfg.db.Model(&User{}).Where("id = ?", user.Id).Updates(map[string]any{"otp_secret": secret, "otp_enable": false}).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(gin.H{
		"otp_secret":  secret,
		"otpauth_url": fmt.Sprintf("otpauth://totp/%s?secret=%s&issuer=e2auth", user.Email, secret),
	}))
}

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
	if user.OTPSecret == "" && !user.OTPEnable {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "MFA not enabled"))
		return
	}
	if !verifyUserMFA(user, input.Token) {
		c.AbortWithStatusJSON(http.StatusUnauthorized, errResp(ErrCodeInvalidCredentials, "Invalid MFA token"))
		return
	}
	if !user.OTPEnable {
		if err := cfg.db.Model(&User{}).Where("id = ?", user.Id).Update("otp_enable", true).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
			return
		}
	}
	c.JSON(http.StatusOK, successResp(nil))
}

func mfaDisable(c *gin.Context) {
	user, err := getCtxUserOrAbort(c)
	if err != nil {
		return
	}
	if err := cfg.db.Model(&User{}).Where("id = ?", user.Id).Updates(map[string]any{"otp_enable": false, "otp_secret": ""}).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	_ = cfg.db.Where("user_id = ?", user.Id).Delete(&OTPRecoveryCode{}).Error
	c.JSON(http.StatusOK, successResp(nil))
}

func linkOAuth(c *gin.Context) {
	var input struct {
		Provider string `json:"provider" binding:"required"`
		Code     string `json:"code" binding:"required"`
	}
	if !bindInput(c, &input) {
		return
	}
	user, err := getCtxUserOrAbort(c)
	if err != nil {
		return
	}
	provider, ok := cfg.oauthProviders[input.Provider]
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "unknown provider"))
		return
	}
	accessToken, refreshToken, err := provider.ExchangeCode(input.Code, "")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	if err := upsertOAuthToken(&OAuth2Token{
		UserId:       user.Id,
		Provider:     input.Provider,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(durationSession),
	}); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	if !slices.Contains(user.OAuth2Providers, input.Provider) {
		providers := append(append([]string{}, user.OAuth2Providers...), input.Provider)
		_ = cfg.db.Model(&User{}).Where("id = ?", user.Id).Update("oauth2_providers", providers).Error
	}
	c.JSON(http.StatusOK, successResp(nil))
}

func unlinkOAuth(c *gin.Context) {
	var input struct {
		Provider string `json:"provider" binding:"required"`
	}
	if !bindInput(c, &input) {
		return
	}
	user, err := getCtxUserOrAbort(c)
	if err != nil {
		return
	}
	remaining := make([]string, 0, len(user.OAuth2Providers))
	for _, p := range user.OAuth2Providers {
		if p != input.Provider {
			remaining = append(remaining, p)
		}
	}
	if len(remaining) == 0 && len(user.PasswordHash) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "cannot unlink the last login method"))
		return
	}
	if err := deleteOAuthToken(user.Id, input.Provider); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	_ = cfg.db.Model(&User{}).Where("id = ?", user.Id).Update("oauth2_providers", remaining).Error
	c.JSON(http.StatusOK, successResp(nil))
}

func generateMFABackupCodes(c *gin.Context) {
	user, err := getCtxUserOrAbort(c)
	if err != nil {
		return
	}
	if !user.OTPEnable {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "MFA not enabled"))
		return
	}
	codes := make([]string, 0, 10)
	for range 10 {
		code, err := e2crypto.RandomString(10)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
			return
		}
		codes = append(codes, code)
	}
	if err := replaceRecoveryCodes(user.Id, codes); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(gin.H{"codes": codes}))
}
