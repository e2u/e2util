package e2auth

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/e2u/e2util/e2app"
	"github.com/gin-gonic/gin"
)

func TestRegisterWithE2DB(t *testing.T) {
	gin.SetMode(gin.TestMode)
	conn := setupTestDB(t)
	r := gin.New()
	Register(r, conn, WithSecretKey([]byte("secret")))

	w := doJSON(t, r, http.MethodPost, "/auth/register", "", map[string]string{
		"username": "moduser",
		"email":    "mod@example.com",
		"password": "Abcdef1!",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("register %d %s", w.Code, w.Body.String())
	}
}

func TestMountProtectsAppRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)
	conn := setupTestDB(t)
	r := gin.New()
	app := &e2app.Context{
		DB: conn,
		App: &e2app.AppConfig{
			ExtraProps: map[string]any{
				"secret_key": base64.StdEncoding.EncodeToString([]byte("secret")),
			},
		},
	}
	Mount(r, app)

	r.GET("/me", Required(), func(c *gin.Context) {
		u, err := CurrentUser(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"name": u.Name})
	})

	w := doJSON(t, r, http.MethodGet, "/me", "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated /me %d", w.Code)
	}

	_ = doJSON(t, r, http.MethodPost, "/auth/register", "", map[string]string{
		"username": "mounted",
		"email":    "mount@example.com",
		"password": "Abcdef1!",
	})
	w = doJSON(t, r, http.MethodPost, "/auth/login", "", map[string]string{
		"email":    "mount@example.com",
		"password": "Abcdef1!",
	})
	var loginResp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &loginResp)
	token, _ := loginResp["token"].(string)
	if token == "" {
		t.Fatalf("login token missing: %s", w.Body.String())
	}

	w = doJSON(t, r, http.MethodGet, "/me", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("/me %d %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("mounted")) {
		t.Fatalf("expected username in body: %s", w.Body.String())
	}
}
