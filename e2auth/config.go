package e2auth

import (
	"gorm.io/gorm"
)

type RouterOption func(*routerConfig)

type routerConfig struct {
	db             *gorm.DB
	tableSchema    string
	logger         Logger
	emailer        Emailer
	captchaService CAPTCHAService
	oauthProviders OAuthProviders
	rateLimiter    RateLimiter
	eventNotifier  EventNotifier
	secretKey      []byte
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
