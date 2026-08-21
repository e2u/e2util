package e2auth

import (
	"context"
	"errors"

	"github.com/e2u/e2util/e2app"
	"github.com/e2u/e2util/e2db"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Register mounts /auth using an e2db connection (read-write) and migrates
// auth tables when EnableMigrate is true.
func Register(router gin.IRouter, conn *e2db.Connect, opts ...RouterOption) {
	if conn == nil {
		panic("e2auth: db connection is nil")
	}
	cfg = newRouterConfig(conn.RW(), opts...)
	if cfg.db == nil {
		panic("e2auth: register failed: db is nil")
	}
	if err := conn.AutoMigrate(context.Background(), authModels()...); err != nil {
		logrus.WithField("model", "e2auth").Errorf("auto migrate failed: %v", err)
	}
	registerAuthRoutes(router)
}

// Mount wires e2auth into an application that already has e2app + e2db.
// Secret comes from app.App "secret_key" (raw or base64) unless WithSecretKey is passed.
func Mount(router gin.IRouter, app *e2app.Context, opts ...RouterOption) {
	if app == nil {
		panic("e2auth: app context is nil")
	}
	if app.DB == nil {
		panic("e2auth: app.DB is nil; configure [orm] in e2app")
	}

	derived := make([]RouterOption, 0, 3+len(opts))
	if app.App != nil {
		if secret := app.App.GetBytesFromBase64("secret_key"); len(secret) > 0 {
			derived = append(derived, WithSecretKey(secret))
		} else if s := app.App.GetString("secret_key"); s != "" {
			derived = append(derived, WithSecretKey([]byte(s)))
		}
	}
	derived = append(derived, WithLogger(logrus.StandardLogger()))
	derived = append(derived, opts...)
	Register(router, app.DB, derived...)
}

// CurrentUserID returns the authenticated user id set by Required/Mount routes.
func CurrentUserID(c *gin.Context) (string, bool) {
	v, ok := c.Get(ctxKeyUserId)
	if !ok {
		return "", false
	}
	id, ok := v.(string)
	return id, ok && id != ""
}

// CurrentUser loads the authenticated user. Call after Required().
func CurrentUser(c *gin.Context) (*User, error) {
	id, ok := CurrentUserID(c)
	if !ok {
		return nil, errors.New("e2auth: no user in context; attach e2auth.Required()")
	}
	return getUserByID(id)
}
