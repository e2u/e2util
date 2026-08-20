package e2auth

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var cfg *routerConfig

func RegisterRouters(router *gin.RouterGroup, db *gorm.DB, opts ...RouterOption) {
	cfg = &routerConfig{
		db:             db,
		tableSchema:    "e2auth",
		logger:         &noopLogger{},
		emailer:        &noopEmailer{},
		captchaService: &noopCAPTCHAService{},
		oauthProviders: make(OAuthProviders),
		rateLimiter:    &noopRateLimiter{},
		eventNotifier:  &noopEventNotifier{},
		secretKey:      []byte("secret key"),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.db == nil {
		panic("e2auth: register router failed: db is nil")
	}

	tables := []schema.Tabler{&User{}, &PasswordResetToken{}, &OTPRecoveryCode{}, &Session{}, &OAuth2Token{}}
	if cfg.db.Name() == "postgres" {
		cfg.db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", cfg.tableSchema))
		for _, table := range tables {
			if err := cfg.db.Table(fmt.Sprintf("%s.%s", cfg.tableSchema, table.TableName())).AutoMigrate(table); err != nil {
				logrus.WithField("model", "e2auth").Errorf("register model failed: %v", err)
			}
		}
	} else {
		for _, table := range tables {
			if err := cfg.db.Table(fmt.Sprintf("%s_%s", cfg.tableSchema, table.TableName())).AutoMigrate(table); err != nil {
				logrus.WithField("model", "e2auth").Errorf("register model failed: %v", err)
			}
		}
	}

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
