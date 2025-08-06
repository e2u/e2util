package e2auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func unlockAccount(c *gin.Context) {
	// Placeholder: Implement account unlock logic using cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Unlock account not implemented"})
}

func oauthProvider(c *gin.Context) {
	// Placeholder: Implement OAuth provider redirect using cfg.oauthProviders
	c.JSON(http.StatusNotImplemented, gin.H{"error": "OAuth provider not implemented"})
}

func oauthCallback(c *gin.Context) {
	// Placeholder: Implement OAuth callback handling using cfg.oauthProviders, cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "OAuth callback not implemented"})
}

func captchaVerify(c *gin.Context) {
	// Placeholder: Implement CAPTCHA verification using cfg.captchaService
	c.JSON(http.StatusNotImplemented, gin.H{"error": "CAPTCHA verify not implemented"})
}

func getProfile(c *gin.Context) {
	// Placeholder: Implement profile retrieval using cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Get profile not implemented"})
}

func putProfile(c *gin.Context) {
	// Placeholder: Implement profile update (full) using cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Put profile not implemented"})
}

func patchProfile(c *gin.Context) {
	// Placeholder: Implement profile update (partial) using cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Patch profile not implemented"})
}

func deleteAccount(c *gin.Context) {
	// Placeholder: Implement account deletion using cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Delete account not implemented"})
}

func mfaEnable(c *gin.Context) {
	// Placeholder: Implement MFA enable logic using cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "MFA enable not implemented"})
}

func mfaVerify(c *gin.Context) {
	// Placeholder: Implement MFA verify logic using cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "MFA verify not implemented"})
}

func mfaDisable(c *gin.Context) {
	// Placeholder: Implement MFA disable logic using cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "MFA disable not implemented"})
}

func getSessions(c *gin.Context) {
	// Placeholder: Implement session listing using cfg.sessoiner
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Get sessions not implemented"})
}

func deleteSession(c *gin.Context) {
	// Placeholder: Implement session deletion using cfg.sessoiner
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Delete session not implemented"})
}

func listUsers(c *gin.Context) {
	// Placeholder: Implement user listing for admins using cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "List users not implemented"})
}

func getUser(c *gin.Context) {
	// Placeholder: Implement user retrieval for admins using cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Get user not implemented"})
}

func updateUser(c *gin.Context) {
	// Placeholder: Implement user update for admins using cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Update user not implemented"})
}

func deleteUser(c *gin.Context) {
	// Placeholder: Implement user deletion for admins using cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Delete user not implemented"})
}

func updateUserRoles(c *gin.Context) {
	// Placeholder: Implement user role update for admins using cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Update user roles not implemented"})
}

func revokeTokens(c *gin.Context) {
	// Placeholder: Implement token revocation using cfg.sessoiner
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Revoke tokens not implemented"})
}

func getProfileRoles(c *gin.Context) {
	// Placeholder: Implement role retrieval using cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Get profile roles not implemented"})
}

func searchUsers(c *gin.Context) {
	// Placeholder: Implement user search using cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Search users not implemented"})
}

func generateMFABackupCodes(c *gin.Context) {
	// Placeholder: Implement backup code generation using cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Generate MFA backup codes not implemented"})
}

func linkOAuth(c *gin.Context) {
	// Placeholder: Implement OAuth linking using cfg.oauthProviders, cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Link OAuth not implemented"})
}

func unlinkOAuth(c *gin.Context) {
	// Placeholder: Implement OAuth unlinking using cfg.storager
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Unlink OAuth not implemented"})
}

func getConfig(c *gin.Context) {
	// Placeholder: Implement config retrieval
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Get config not implemented"})
}

func updateConfig(c *gin.Context) {
	// Placeholder: Implement config update
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Update config not implemented"})
}

// func getCSRFToken(c *gin.Context) {
//	userId, exists := c.Get(ctxKeyUserId)
//	if !exists {
//		c.JSON(http.StatusUnauthorized, errResp(ErrCodeUnauthorized, "Unauthorized"))
//		return
//	}
//	sessionToken := c.GetHeader("Authorization")
//	subject, err := e2jwt.VerifyWithEncryptSubject[*SessionSubject](sessionToken, cfg.storager.GetSecretKey())
//	if err != nil {
//		c.JSON(http.StatusUnauthorized, errResp(ErrCodeUnauthorized, err))
//		return
//	}
//
//	if subject.UserId != userId.(string) {
//		c.JSON(http.StatusUnauthorized, errResp(ErrCodeUnauthorized, "Unauthorized"))
//		return
//	}
//
//	csrfToken, err := generateCSRFToken(uuid.NewString(), subject.SessionId, durationCSRF, cfg)
//	if err != nil {
//		c.JSON(http.StatusUnauthorized, errResp(ErrCodeUnauthorized, err))
//		return
//	}
//	c.JSON(http.StatusOK, gin.H{"csrf_token": csrfToken})
//}
