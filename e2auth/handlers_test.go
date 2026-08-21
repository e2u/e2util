package e2auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/e2u/e2util/e2crypto"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type recEmailer struct {
	mu   sync.Mutex
	to   string
	data map[string]any
}

func (e *recEmailer) SendEmail(to, subject, body string) error { return nil }
func (e *recEmailer) SendTemplateEmail(to string, data map[string]any) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.to = to
	e.data = data
	return nil
}

type mockOAuth struct{}

func (m mockOAuth) GetAuthURL(state string) string { return "https://example.com/oauth?state=" + state }
func (m mockOAuth) ExchangeCode(code, redirectURI string) (string, string, error) {
	return "access-" + code, "refresh-" + code, nil
}
func (m mockOAuth) GetUserInfo(accessToken string) (map[string]any, error) {
	return map[string]any{"email": "oauth@example.com", "id": "oid-1", "name": "oauth-user"}, nil
}

func setupAuthEngine(t *testing.T, mail *recEmailer) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	conn := setupTestDB(t)
	if mail == nil {
		mail = &recEmailer{}
	}
	r := gin.New()
	RegisterRouters(r.Group("/"), conn.RW(),
		WithSecretKey([]byte("secret")),
		WithEmailer(mail),
		WithOAuthProviders(OAuthProviders{"mock": mockOAuth{}}),
	)
	return r
}

func doJSON(t *testing.T, r http.Handler, method, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRegisterLoginVerifyToken(t *testing.T) {
	r := setupAuthEngine(t, nil)
	w := doJSON(t, r, http.MethodPost, "/auth/register", "", map[string]string{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "Abcdef1!",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("register %d %s", w.Code, w.Body.String())
	}

	w = doJSON(t, r, http.MethodPost, "/auth/login", "", map[string]string{
		"email":    "alice@example.com",
		"password": "Abcdef1!",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("login %d %s", w.Code, w.Body.String())
	}
	var loginResp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &loginResp); err != nil {
		t.Fatal(err)
	}
	token, _ := loginResp["token"].(string)
	if token == "" {
		t.Fatalf("missing token: %s", w.Body.String())
	}

	w = doJSON(t, r, http.MethodPost, "/auth/verify-token", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("verify-token %d %s", w.Code, w.Body.String())
	}
}

func TestEmailVerifyStoresAndChecksCode(t *testing.T) {
	mail := &recEmailer{}
	r := setupAuthEngine(t, mail)
	_ = doJSON(t, r, http.MethodPost, "/auth/register", "", map[string]string{
		"username": "bob",
		"email":    "bob@example.com",
		"password": "Abcdef1!",
	})
	w := doJSON(t, r, http.MethodPost, "/auth/login", "", map[string]string{
		"email":    "bob@example.com",
		"password": "Abcdef1!",
	})
	var loginResp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &loginResp)
	token, _ := loginResp["token"].(string)

	w = doJSON(t, r, http.MethodPost, "/auth/verify-email/send", token, map[string]string{})
	if w.Code != http.StatusOK {
		t.Fatalf("send %d %s", w.Code, w.Body.String())
	}
	code, _ := mail.data["code"].(string)
	if code == "" {
		t.Fatal("email code not sent")
	}

	w = doJSON(t, r, http.MethodPost, "/auth/verify-email/confirm", token, map[string]string{"code": "000000"})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad code should 400, got %d", w.Code)
	}
	w = doJSON(t, r, http.MethodPost, "/auth/verify-email/confirm", token, map[string]string{"code": code})
	if w.Code != http.StatusOK {
		t.Fatalf("confirm %d %s", w.Code, w.Body.String())
	}
}

func TestAccountLockAndUnlock(t *testing.T) {
	mail := &recEmailer{}
	r := setupAuthEngine(t, mail)
	_ = doJSON(t, r, http.MethodPost, "/auth/register", "", map[string]string{
		"username": "locked",
		"email":    "lock@example.com",
		"password": "Abcdef1!",
	})
	for i := 0; i < maxLoginFailures; i++ {
		w := doJSON(t, r, http.MethodPost, "/auth/login", "", map[string]string{
			"email":    "lock@example.com",
			"password": "WrongPass1!",
		})
		if w.Code != http.StatusUnauthorized && w.Code != http.StatusForbidden {
			t.Fatalf("attempt %d: %d %s", i, w.Code, w.Body.String())
		}
	}
	w := doJSON(t, r, http.MethodPost, "/auth/login", "", map[string]string{
		"email":    "lock@example.com",
		"password": "Abcdef1!",
	})
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected locked 403, got %d %s", w.Code, w.Body.String())
	}

	w = doJSON(t, r, http.MethodPost, "/auth/unlock-account", "", map[string]string{"email": "lock@example.com"})
	if w.Code != http.StatusOK {
		t.Fatalf("unlock send %d %s", w.Code, w.Body.String())
	}
	tok, _ := mail.data["token"].(string)
	if tok == "" {
		t.Fatal("unlock token not emailed")
	}
	w = doJSON(t, r, http.MethodPost, "/auth/unlock-account", "", map[string]string{
		"email": "lock@example.com",
		"token": tok,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("unlock confirm %d %s", w.Code, w.Body.String())
	}
	w = doJSON(t, r, http.MethodPost, "/auth/login", "", map[string]string{
		"email":    "lock@example.com",
		"password": "Abcdef1!",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("login after unlock %d %s", w.Code, w.Body.String())
	}
}

func TestMFABackupCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	conn := setupTestDB(t)
	cfg = &routerConfig{
		db:             conn.RW(),
		logger:         &noopLogger{},
		emailer:        &noopEmailer{},
		captchaService: &noopCAPTCHAService{},
		oauthProviders: make(OAuthProviders),
		rateLimiter:    &noopRateLimiter{},
		eventNotifier:  &noopEventNotifier{},
		secretKey:      []byte("secret"),
	}
	secret, err := e2crypto.GenerateTOTPSecret()
	if err != nil {
		t.Fatal(err)
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("Abcdef1!"), bcrypt.DefaultCost)
	user := &User{Id: "mfa-user", Name: "mfauser", Email: "mfa@example.com", OTPEnable: true, OTPSecret: secret, PasswordHash: hash}
	if err := createUser(user); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(ctxKeyUserId, user.Id)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/mfa/backup-codes", nil)
	generateMFABackupCodes(c)
	if w.Code != http.StatusOK {
		t.Fatalf("backup codes %d %s", w.Code, w.Body.String())
	}
	var resp struct {
		Codes []string `json:"codes"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Codes) != 10 {
		t.Fatalf("got %d codes", len(resp.Codes))
	}

	body, _ := json.Marshal(map[string]string{"token": resp.Codes[0]})
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set(ctxKeyUserId, user.Id)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/mfa/verify", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	mfaVerify(c)
	if w.Code != http.StatusOK {
		t.Fatalf("backup verify %d %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Set(ctxKeyUserId, user.Id)
	c.Request = httptest.NewRequest(http.MethodPost, "/auth/mfa/verify", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	mfaVerify(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("reused backup code should 401, got %d", w.Code)
	}
}

func TestOAuthCallbackCreatesSession(t *testing.T) {
	r := setupAuthEngine(t, nil)
	w := doJSON(t, r, http.MethodGet, "/auth/oauth/mock/callback?code=abc", "", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("oauth callback %d %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["token"] == nil {
		t.Fatalf("expected session token, got %s", w.Body.String())
	}
}

func TestChangeEmailUsesNewAddress(t *testing.T) {
	mail := &recEmailer{}
	r := setupAuthEngine(t, mail)
	_ = doJSON(t, r, http.MethodPost, "/auth/register", "", map[string]string{
		"username": "carol",
		"email":    "carol@example.com",
		"password": "Abcdef1!",
	})
	w := doJSON(t, r, http.MethodPost, "/auth/login", "", map[string]string{
		"email":    "carol@example.com",
		"password": "Abcdef1!",
	})
	var loginResp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &loginResp)
	token, _ := loginResp["token"].(string)

	w = doJSON(t, r, http.MethodPost, "/auth/change-email", token, map[string]string{
		"password":      "Abcdef1!",
		"current_email": "carol@example.com",
		"new_email":     "carol2@example.com",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("change-email %d %s", w.Code, w.Body.String())
	}
	code, _ := mail.data["code"].(string)
	chgTok, _ := mail.data["change_email_token"].(string)
	w = doJSON(t, r, http.MethodPost, "/auth/change-email/confirm", token, map[string]string{
		"code":               code,
		"change_email_token": chgTok,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("confirm %d %s", w.Code, w.Body.String())
	}

	w = doJSON(t, r, http.MethodPost, "/auth/login", "", map[string]string{
		"email":    "carol2@example.com",
		"password": "Abcdef1!",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("login new email %d %s", w.Code, w.Body.String())
	}
}

func TestLoginRequiresMFAToken(t *testing.T) {
	r := setupAuthEngine(t, nil)
	secret, err := e2crypto.GenerateTOTPSecret()
	if err != nil {
		t.Fatal(err)
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte("Abcdef1!"), bcrypt.DefaultCost)
	if err := cfg.db.Create(&User{
		Id: "mfa-login", Name: "mfalogin", Email: "mfalogin@example.com",
		PasswordHash: hash, OTPEnable: true, OTPSecret: secret, ExternalID: "local:mfa-login",
	}).Error; err != nil {
		t.Fatal(err)
	}
	w := doJSON(t, r, http.MethodPost, "/auth/login", "", map[string]string{
		"email":    "mfalogin@example.com",
		"password": "Abcdef1!",
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected mfa required, got %d %s", w.Code, w.Body.String())
	}
	code, err := e2crypto.GenerateTOTP(secret, time.Now(), 30)
	if err != nil {
		t.Fatal(err)
	}
	w = doJSON(t, r, http.MethodPost, "/auth/login", "", map[string]string{
		"email":     "mfalogin@example.com",
		"password":  "Abcdef1!",
		"mfa_token": code,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("login with mfa %d %s", w.Code, w.Body.String())
	}
}
