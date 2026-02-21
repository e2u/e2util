package e2auth

import (
    "bytes"
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/e2u/e2util/e2crypto"
    "github.com/e2u/e2util/e2db"
    "github.com/e2u/e2util/e2logrus"
    "github.com/gin-gonic/gin"
)

func TestMFAVerifySuccess(t *testing.T) {
    // Setup test DB
    conn := setupTestDB(t)

    // Initialize global cfg required by handlers
    cfg = &routerConfig{
        db:            conn.RW(),
        logger:        &noopLogger{},
        emailer:       &noopEmailer{},
        captchaService: &noopCAPTCHAService{},
        oauthProviders: make(OAuthProviders),
        rateLimiter:   &noopRateLimiter{},
        eventNotifier: &noopEventNotifier{},
        secretKey:    []byte("secret"),
    }

    // Auto‑migrate tables needed for the test
    if err := conn.RW().AutoMigrate(&User{}, &Session{}, &PasswordResetToken{}, &OTPRecoveryCode{}, &OAuth2Token{}); err != nil {
        t.Fatalf("auto migrate failed: %v", err)
    }

    // Create a user with MFA enabled using a valid base32 secret
    secret, _ := e2crypto.GenerateTOTPSecret()
    now := time.Now()
    user := &User{
        Id:         "test-id",
        Name:       "test",
        Email:      "test1@example.com",
        ExternalID: "ext1",
        OTPEnable:  true,
        OTPSecret:  secret,
        LastLogin:  now,
    }
    if err := createUser(user); err != nil {
        t.Fatalf("create user failed: %v", err)
    }

    // Generate a valid TOTP code
    validCode := e2crypto.GenerateTOTP(secret, time.Now(), 30)

    // Build request
    body, _ := json.Marshal(map[string]string{"token": validCode})
    req, _ := http.NewRequest(http.MethodPost, "/auth/mfa/verify", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Set(ctxKeyUserId, user.Id)

    // Call handler
    mfaVerify(c)

    if w.Code != http.StatusOK {
        t.Fatalf("expected 200 OK, got %d", w.Code)
    }
}

func TestMFAVerifyFailure(t *testing.T) {
    conn := setupTestDB(t)

    cfg = &routerConfig{
        db:            conn.RW(),
        logger:        &noopLogger{},
        emailer:       &noopEmailer{},
        captchaService: &noopCAPTCHAService{},
        oauthProviders: make(OAuthProviders),
        rateLimiter:   &noopRateLimiter{},
        eventNotifier: &noopEventNotifier{},
        secretKey:    []byte("secret"),
    }

    if err := conn.RW().AutoMigrate(&User{}, &Session{}, &PasswordResetToken{}, &OTPRecoveryCode{}, &OAuth2Token{}); err != nil {
        t.Fatalf("auto migrate failed: %v", err)
    }

    // Create a user with MFA enabled using a valid base32 secret
    secret, _ := e2crypto.GenerateTOTPSecret()
    now := time.Now()
    user := &User{
        Id:         "test-id-2",
        Name:       "test2",
        Email:      "test2@example.com",
        ExternalID: "ext2",
        OTPEnable:  true,
        OTPSecret:  secret,
        LastLogin:  now,
    }
    if err := createUser(user); err != nil {
        t.Fatalf("create user failed: %v", err)
    }

    // Use an invalid TOTP code
    body, _ := json.Marshal(map[string]string{"token": "000000"})
    req, _ := http.NewRequest(http.MethodPost, "/auth/mfa/verify", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Set(ctxKeyUserId, user.Id)

    mfaVerify(c)

    if w.Code != http.StatusUnauthorized {
        t.Fatalf("expected 401 Unauthorized, got %d", w.Code)
    }
}

// Helper to set up test DB (mirrors e2db's test setup)
func setupTestDB(t *testing.T) *e2db.Connect {
    cfg := &e2db.Config{
        Driver:        "sqlite",
        Writer:        "file::memory:?cache=shared",
        Readers:       []string{"file::memory:?cache=shared"},
        LoggerConfig:  &e2logrus.Config{Level: "info", Format: "text"},
    }
    conn, err := e2db.New(cfg)
    if err != nil {
        t.Fatalf("failed to create test DB: %v", err)
    }
    if err := conn.AutoMigrate(context.Background(), &User{}, &Session{}, &PasswordResetToken{}, &OTPRecoveryCode{}, &OAuth2Token{}); err != nil {
        t.Fatalf("auto migrate failed: %v", err)
    }
    return conn
}
