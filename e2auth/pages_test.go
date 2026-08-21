package e2auth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/e2u/e2util/e2gin"
	"github.com/gin-gonic/gin"
)

func TestAuthPagesRender(t *testing.T) {
	r := setupAuthEngine(t, nil)
	for _, path := range []string{"/auth/login", "/auth/register", "/auth/forgot-password"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Accept", "text/html")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("%s -> %d %s", path, w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "<form") {
			t.Fatalf("%s missing form", path)
		}
	}
}

func TestHTMLLoginSetsCookieAndOpensAccount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := setupAuthEngine(t, nil)
	_ = doJSON(t, r, http.MethodPost, "/auth/register", "", map[string]string{
		"username": "pageuser",
		"email":    "page@example.com",
		"password": "Abcdef1!",
	})

	form := url.Values{}
	form.Set("email", "page@example.com")
	form.Set("password", "Abcdef1!")
	form.Set("next", "/auth/account")
	req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusSeeOther {
		t.Fatalf("form login %d %s", w.Code, w.Body.String())
	}
	cookie := ""
	for _, c := range w.Result().Cookies() {
		if c.Name == sessionCookie {
			cookie = c.Value
		}
	}
	if cookie == "" {
		t.Fatal("missing session cookie")
	}

	req = httptest.NewRequest(http.MethodGet, "/auth/account", nil)
	req.Header.Set("Accept", "text/html")
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: cookie})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("account %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "page@example.com") {
		t.Fatalf("account body: %s", w.Body.String())
	}
}

func TestGinTemplatesRenderLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	conn := setupTestDB(t)
	r := gin.New()
	tmpl, err := e2gin.ParseTemplates(TemplateFS())
	if err != nil {
		t.Fatal(err)
	}
	r.SetHTMLTemplate(tmpl)
	Register(r, conn, WithSecretKey([]byte("secret")), WithGinTemplates())

	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	req.Header.Set("Accept", "text/html")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("gin template login %d %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `<form method="post" action="/auth/login">`) {
		t.Fatalf("missing form: %s", w.Body.String())
	}
}

func TestAuthPagesLanguageSwitch(t *testing.T) {
	r := setupAuthEngine(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/auth/login?lang=en", nil)
	req.Header.Set("Accept", "text/html")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("en login %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, ">Sign in<") {
		t.Fatalf("expected English Sign in: %s", body)
	}
	if strings.Contains(body, ">登入<") {
		t.Fatalf("Chinese label should not be primary in en: %s", body)
	}
	langCookieVal := ""
	for _, c := range w.Result().Cookies() {
		if c.Name == langCookie {
			langCookieVal = c.Value
		}
	}
	if langCookieVal != "en" {
		t.Fatalf("lang cookie = %q", langCookieVal)
	}

	req = httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	req.Header.Set("Accept", "text/html")
	req.AddCookie(&http.Cookie{Name: langCookie, Value: "en"})
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if !strings.Contains(w.Body.String(), ">Sign in<") {
		t.Fatalf("cookie should keep English: %s", w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/auth/login?lang=zh", nil)
	req.Header.Set("Accept", "text/html")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if !strings.Contains(w.Body.String(), ">登入<") {
		t.Fatalf("expected Chinese 登入: %s", w.Body.String())
	}
}

func TestAccountPageRedirectsWhenAnonymous(t *testing.T) {
	r := setupAuthEngine(t, nil)
	req := httptest.NewRequest(http.MethodGet, "/auth/account", nil)
	req.Header.Set("Accept", "text/html")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusFound {
		t.Fatalf("expected redirect, got %d", w.Code)
	}
}
