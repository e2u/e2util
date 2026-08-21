package e2auth

import (
	"gorm.io/gorm"
)

type RouterOption func(*routerConfig)

type routerConfig struct {
	db              *gorm.DB
	tableSchema     string
	logger          Logger
	emailer         Emailer
	captchaService  CAPTCHAService
	oauthProviders  OAuthProviders
	rateLimiter     RateLimiter
	eventNotifier   EventNotifier
	secretKey       []byte
	disablePages    bool
	appName         string
	useGinTemplates bool
}

func WithSecretKey(secretKey []byte) RouterOption {
	return func(config *routerConfig) {
		config.secretKey = secretKey
	}
}

func WithDB(db *gorm.DB) RouterOption {
	return func(r *routerConfig) {
		r.db = db
	}
}

func WithTableSchema(schema string) RouterOption {
	return func(r *routerConfig) {
		r.tableSchema = schema
	}
}
func WithLogger(logger Logger) RouterOption {
	return func(cfg *routerConfig) {
		cfg.logger = logger
	}
}

func WithEmailer(emailer Emailer) RouterOption {
	return func(cfg *routerConfig) {
		cfg.emailer = emailer
	}
}

func WithCAPTCHAService(captchaService CAPTCHAService) RouterOption {
	return func(cfg *routerConfig) {
		cfg.captchaService = captchaService
	}
}

func WithOAuthProviders(oauthProviders OAuthProviders) RouterOption {
	return func(cfg *routerConfig) {
		cfg.oauthProviders = oauthProviders
	}
}

func WithRateLimiter(rateLimiter RateLimiter) RouterOption {
	return func(cfg *routerConfig) {
		cfg.rateLimiter = rateLimiter
	}
}

func WithEventNotifier(eventNotifier EventNotifier) RouterOption {
	return func(cfg *routerConfig) {
		cfg.eventNotifier = eventNotifier
	}
}

// WithDisablePages skips the built-in login/register HTML pages (API-only).
func WithDisablePages() RouterOption {
	return func(cfg *routerConfig) {
		cfg.disablePages = true
	}
}

// WithAppName sets the title shown on HTML auth pages.
func WithAppName(name string) RouterOption {
	return func(cfg *routerConfig) {
		cfg.appName = name
	}
}

// WithGinTemplates renders auth pages via gin HTML templates (e2gin).
// Put login.html, register.html, forgot.html, reset.html, account.html
// on the engine (see TemplateFS), then c.HTML uses those names.
func WithGinTemplates() RouterOption {
	return func(cfg *routerConfig) {
		cfg.useGinTemplates = true
	}
}
