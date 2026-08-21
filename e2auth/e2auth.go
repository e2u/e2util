package e2auth

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var cfg *routerConfig

func authModels() []any {
	return []any{&User{}, &PasswordResetToken{}, &OTPRecoveryCode{}, &Session{}, &OAuth2Token{}, &Configuration{}}
}

func newRouterConfig(db *gorm.DB, opts ...RouterOption) *routerConfig {
	c := &routerConfig{
		db:             db,
		tableSchema:    "e2auth",
		logger:         &noopLogger{},
		emailer:        &noopEmailer{},
		captchaService: &noopCAPTCHAService{},
		oauthProviders: make(OAuthProviders),
		rateLimiter:    &noopRateLimiter{},
		eventNotifier:  &noopEventNotifier{},
		secretKey:      []byte("secret key"),
		appName:        "Account",
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func migrateAuthTables(db *gorm.DB) {
	if db.Name() == "postgres" && cfg != nil && cfg.tableSchema != "" {
		if err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", cfg.tableSchema)).Error; err != nil {
			logrus.WithField("model", "e2auth").Warnf("create schema %s: %v", cfg.tableSchema, err)
		}
	}
	for _, table := range authModels() {
		if err := db.AutoMigrate(table); err != nil {
			logrus.WithField("model", "e2auth").Errorf("register model failed: %v", err)
		}
	}
}

// RegisterRouters mounts /auth onto a Gin router using a Gorm DB.
// Prefer Register (e2db) or Mount (e2app) when wiring a full application.
func RegisterRouters(router gin.IRouter, db *gorm.DB, opts ...RouterOption) {
	cfg = newRouterConfig(db, opts...)
	if cfg.db == nil {
		panic("e2auth: register router failed: db is nil")
	}
	migrateAuthTables(cfg.db)
	registerAuthRoutes(router)
}

func registerAuthRoutes(router gin.IRouter) {
	auth := router.Group("/auth")
	auth.Use(loggingMiddleware(cfg.logger))
	{
		/* done */ auth.POST("/login", rateLimitMiddleware(cfg.rateLimiter, 5, time.Minute), login)
		/* done */ auth.POST("/register", rateLimitMiddleware(cfg.rateLimiter, 3, time.Minute), register)
		/* done */ auth.POST("/reset-password", rateLimitMiddleware(cfg.rateLimiter, 3, time.Minute), resetPassword)
		/* done */ auth.POST("/reset-password/confirm", rateLimitMiddleware(cfg.rateLimiter, 3, time.Minute), resetPasswordConfirm)
		/* done */ auth.GET("/logout", logout)
		auth.POST("/unlock-account", unlockAccount)
		auth.GET("/oauth/:provider", oauthProvider)
		auth.GET("/oauth/:provider/callback", oauthCallback)
		auth.POST("/captcha/verify", captchaVerify)
		if cfg == nil || !cfg.disablePages {
			auth.GET("/login", pageLogin)
			auth.GET("/register", pageRegister)
			auth.GET("/forgot-password", pageForgot)
			auth.GET("/reset-password", pageReset)
			auth.GET("/account", pageAccount)
		}
	}

	protected := auth.Group("")
	protected.Use(authMiddleware())
	{
		protected.POST("/verify-email/send", verifyEmailSent)
		protected.POST("/verify-email/confirm", verifyEmailConfirm)
		protected.POST("/refresh-token", refreshToken)
		protected.POST("/verify-token", verifyToken)
		/* done */
		protected.POST("/change-password", changePassword)
		/* done */
		protected.POST("/change-email", changeEmail)
		/* done */
		protected.POST("/change-email/confirm", changeEmailConfirm)

		protected.GET("/profile", func(c *gin.Context) {
			getProfile(c)
		})
		protected.PUT("/profile", func(c *gin.Context) {
			putProfile(c)
		})
		protected.PATCH("/profile", func(c *gin.Context) {
			patchProfile(c)
		})
		protected.POST("/profile/oauth/link", func(c *gin.Context) {
			linkOAuth(c)
		})
		protected.POST("/profile/oauth/unlink", func(c *gin.Context) {
			unlinkOAuth(c)
		})
		protected.DELETE("/account", func(c *gin.Context) {
			deleteAccount(c)
		})
		protected.POST("/mfa/enable", func(c *gin.Context) {
			mfaEnable(c)
		})
		protected.POST("/mfa/verify", func(c *gin.Context) {
			mfaVerify(c)
		})
		protected.POST("/mfa/disable", func(c *gin.Context) {
			mfaDisable(c)
		})
		protected.GET("/sessions", func(c *gin.Context) {
			getSessions(c)
		})
		protected.DELETE("/sessions/:id", func(c *gin.Context) {
			deleteSession(c)
		})
		protected.POST("/revoke-tokens", func(c *gin.Context) {
			revokeTokens(c)
		})
		protected.GET("/profile/roles", func(c *gin.Context) {
			getProfileRoles(c)
		})
		protected.POST("/mfa/backup-codes", func(c *gin.Context) {
			generateMFABackupCodes(c)
		})
	}

	admin := auth.Group("/admin")
	admin.Use(authMiddleware())
	admin.Use(adminMiddleware())
	{
		admin.GET("/users", func(c *gin.Context) {
			listUsers(c)
		})
		admin.GET("/users/:id", func(c *gin.Context) {
			getUser(c)
		})
		admin.PUT("/users/:id", func(c *gin.Context) {
			updateUser(c)
		})
		admin.DELETE("/users/:id", func(c *gin.Context) {
			deleteUser(c)
		})
		admin.PUT("/users/:id/roles", func(c *gin.Context) {
			updateUserRoles(c)
		})
		admin.GET("/users/search", func(c *gin.Context) {
			searchUsers(c)
		})

		admin.GET("/config", func(c *gin.Context) {
			getConfig(c)
		})
		admin.PUT("/config", func(c *gin.Context) {
			updateConfig(c)
		})
	}
}
