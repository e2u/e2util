# e2auth

同 `e2app`、`e2db`、`e2gin` 一齊用嘅認證模組。插上去就有註冊、登入、重設密碼、電郵驗證、MFA、會話、OAuth，同內建中英 HTML 頁。業務應用唔使再自己起用戶帳號系統。

Auth module for use with `e2app`, `e2db`, and `e2gin`. Drop it in for register, login, password reset, email verify, MFA, sessions, OAuth, and built-in Chinese/English HTML pages. Applications do not need to reimplement accounts.

電郵、CAPTCHA、限流、OAuth provider 預設為 no-op；上線請注入真實實作。

Email, CAPTCHA, rate limiting, and OAuth providers default to no-ops. Inject real implementations for production.

---

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2auth
```

依賴：`e2app`（設定 + DB）、`e2db`（連線／遷移）、`e2gin`（HTTP）。亦可只用 `e2db` + Gin，見下「只接 e2db」。

Depends on `e2app` (config + DB), `e2db` (connect/migrate), and `e2gin` (HTTP). You can also wire `e2db` + Gin only; see “Register with e2db only”.

---

## 整體接法 / How the pieces fit

| 套件 / Package | 職責 / Role |
| --- | --- |
| `e2app` | 讀 TOML／環境變數，提供 `app.DB`、`secret_key`、HTTP 位址 |
| `e2db` | Postgres／SQLite 連線；`enable_migrate = true` 時建 auth 表 |
| `e2gin` | `DefaultEngine`、靜態檔、模板熱重載、`StartAndStopHttp` |
| `e2auth` | `/auth/*` API + HTML 頁；`Required()` 保護你自己嘅路由 |

建議流程 / Recommended flow:

1. `e2app.New(...)` 載入設定同 DB  
2. `e2gin.DefaultEngine(...)` 建立引擎  
3. `e2auth.Mount(eng, app)` 掛認證  
4. 用 `e2auth.Required()` 保護業務路由  
5. `e2gin.StartAndStopHttp(...)` 開伺服器  

---

## 設定檔 / Configuration

`Mount` 需要 `[app].secret_key`（簽 session JWT）同 `[orm]`。`enable_migrate = true` 時會自動建表：`users`、`sessions`、`password_reset_tokens`、`otp_recovery_codes`、`oauth2_tokens`、`configurations`。

`Mount` needs `[app].secret_key` (JWT) and `[orm]`. With `enable_migrate = true` it creates the auth tables.

```toml
[app]
name = "my-app"
secret_key = "c2VjcmV0X2tleQo="   # base64，或直接寫明文 / base64 or raw string

[orm]
driver = "postgres"               # 或 sqlite / or sqlite
writer = "host=127.0.0.1 port=5432 user=pgsql password=123456 dbname=myapp sslmode=disable"
enable_migrate = true

[http]
address = "0.0.0.0"
port = 8080
```

`secret_key`：`GetBytesFromBase64("secret_key")` 成功就用解碼結果，否則用原始字串。請用足夠長嘅密鑰，唔好用預設 `"secret key"` 上線。

`secret_key` is decoded as base64 when valid, otherwise used as a raw string. Use a long secret in production; do not ship the default `"secret key"`.

---

## 最小例子 / Minimal example

```go
package main

import (
	"context"
	"embed"

	"github.com/e2u/e2util/e2app"
	"github.com/e2u/e2util/e2auth"
	"github.com/e2u/e2util/e2gin"
	"github.com/gin-gonic/gin"
)

//go:embed *.toml
var cfgFS embed.FS

func main() {
	app := e2app.New(context.Background(), cfgFS)
	eng := e2gin.DefaultEngine(&e2gin.Option{})

	e2auth.Mount(eng, app,
		e2auth.WithAppName(app.App.Name),
		// e2auth.WithEmailer(mailer), // 上線寄重設信 / send reset mail in production
	)

	eng.GET("/api/me", e2auth.Required(), func(c *gin.Context) {
		u, err := e2auth.CurrentUser(c)
		if err != nil {
			c.JSON(401, gin.H{"error": "unauthorized"})
			return
		}
		c.JSON(200, gin.H{"id": u.Id, "email": u.Email, "name": u.Name})
	})

	addr, port := "0.0.0.0", 8080
	if app.Http != nil {
		addr, port = app.Http.Address, app.Http.Port
	}
	e2gin.StartAndStopHttp(eng, addr, port, func() {})
}
```

啟動後瀏覽器開 `http://127.0.0.1:8080/auth/login` 即可註冊／登入，唔使自己寫頁。

After start, open `http://127.0.0.1:8080/auth/login` — no custom account UI required.

---

## 只接 e2db / Register with e2db only

冇用 `e2app.New` 時：

```go
e2auth.Register(engine, dbConnect,
	e2auth.WithSecretKey([]byte("your-long-secret")),
	e2auth.WithAppName("my-app"),
	e2auth.WithEmailer(mailer),
)
```

`Register` 用 `e2db.Connect.RW()`，並喺 `EnableMigrate` 為 true 時遷移表。

`Register` uses `e2db.Connect.RW()` and migrates when `EnableMigrate` is true.

舊 API `RegisterRouters(router, *gorm.DB, opts...)` 仍然可用，新代碼請用 `Mount` 或 `Register`。

---

## 內建 HTML 頁 / Built-in pages

`Mount`／`Register` 預設會掛以下 **GET** 頁（表單 POST 去同一套 JSON API）。預設用 **e2html** 組 HTML，唔使 e2gin 模板。

By default these **GET** pages are mounted (forms POST to the JSON API). Defaults are built with **e2html**, no e2gin templates required.

| Path | 說明 / Meaning |
| --- | --- |
| `GET /auth/login` | 登入 / Sign in |
| `GET /auth/register` | 註冊 / Register |
| `GET /auth/forgot-password` | 忘記密碼 / Forgot password |
| `GET /auth/reset-password?token=` | 設新密碼（郵件連結）/ New password |
| `GET /auth/account` | 帳號摘要 / Account |
| `GET /auth/logout` | 登出並清 cookie / Sign out |

只要 JSON API、唔要頁：

```go
e2auth.Mount(eng, app, e2auth.WithDisablePages())
```

頁面標題用 `WithAppName("我的應用")`。

### 語言 / Language

每頁有 **中文 | English**。

1. `?lang=zh` 或 `?lang=en`（會寫 cookie `e2auth_lang`，有效期一年）  
2. 否則讀 cookie  
3. 再否則 `Accept-Language`（`en*` → 英文）  
4. 預設中文  

切換連結會保留而家嘅 path／query（例如 `next`、`token`）。

### Cookie 同 session

| Cookie | 用途 / Purpose |
| --- | --- |
| `e2auth_session` | HttpOnly session JWT；登入／OAuth 成功時寫入 |
| `e2auth_lang` | 介面語言 `zh` / `en` |

認證順序：先 `Authorization: Bearer <token>`，冇 header 再用 cookie。瀏覽器表單登入會 redirect；JSON 客戶端仍然拿 `token` 欄位。

Auth order: `Authorization: Bearer <token>`, then the session cookie. HTML form login redirects; JSON clients still receive `token`.

HTML GET 未登入時，`Required()` 會 302 去 `/auth/login?next=...`；API 請求回 JSON 401。

Unauthenticated HTML GET + `Required()` redirects to login; API calls get JSON 401.

`next` 只接受以 `/` 開頭、唔係 `//` 嘅相對路徑，避免 open redirect。

### 用 e2gin 模板改版面 / Override with e2gin templates

```go
eng := e2gin.DefaultEngine(&e2gin.Option{
	Template: &e2gin.Template{
		FS:        e2auth.TemplateFS(), // 或同你自己嘅 embed.FS 合併 / or merge with your FS
		LocalPath: "./templates",       // debug 模式熱重載 / hot reload in debug
	},
})
e2auth.Mount(eng, app,
	e2auth.WithGinTemplates(),
	e2auth.WithAppName("我的應用"),
)
```

`e2auth.TemplateFS()` 檔名：

- `login.html`
- `register.html`
- `forgot.html`
- `reset.html`
- `account.html`

可複製到自己嘅 `templates/` 改 CSS／結構。模板資料：

| 欄位 / Field | 說明 / Meaning |
| --- | --- |
| `Title` | `<title>` |
| `AppName` | 應用名 |
| `Next` | 登入後跳轉 |
| `Notice` / `Error` | 提示／錯誤 |
| `Token` | 重設密碼 token |
| `User` | `*e2auth.User`（帳號頁） |
| `Lang` | `zh` 或 `en` |
| `HTMLLang` | `zh-Hant` 或 `en` |
| `T` | 文案：`T.SignIn`、`T.Email`、`T.PasswordHint` 等 |
| `LangURLZH` / `LangURLEN` | 切換語言連結 |

Gin 模板名稱要同檔名一致，例如 `login.html`。

---

## 保護業務路由 / Protect your routes

| 函數 / Function | 說明 / Meaning |
| --- | --- |
| `Required()` | 要有效 session |
| `AdminOnly()` | 要 `roles` 含 `admin`（放喺 `Required()` 之後） |
| `RequireAdmin()` | 認證 + admin 一次過 |
| `CurrentUser(c)` | `(*User, error)` |
| `CurrentUserID(c)` | `(string, bool)` |

```go
api := eng.Group("/api")
api.Use(e2auth.Required())
api.GET("/orders", listOrders)

admin := eng.Group("/api/admin", e2auth.RequireAdmin())
admin.GET("/stats", adminStats)
```

用戶主鍵用 `e2auth.User.Id`（UUID 字串）。業務表用 `user_id` 指向佢，唔好另外起登入用用戶表。

Use `e2auth.User.Id` (UUID string) as the account id. Point business `user_id` columns at it; do not create a second login user table.

管理員：把該用戶 `roles` 設為含 `"admin"`（可用 `PUT /auth/admin/users/:id/roles`，需已係 admin）。

---

## JSON API

`Content-Type: application/json`。成功：`{"status":"success", ...}`。失敗：`{"status":"error","err_code":"...","err_message":"..."}`。

### 密碼規則 / Password rules

至少 8 位，必須含：大寫、小寫、數字、同 `!@#$%^&*` 其中一個。

At least 8 characters, with uppercase, lowercase, a digit, and one of `!@#$%^&*`.

### 公開 / Public

**註冊 / Register** `POST /auth/register`

```json
{"username":"ada","email":"ada@example.com","password":"Abcdef1!"}
```

**登入 / Login** `POST /auth/login`

```json
{"email":"ada@example.com","password":"Abcdef1!"}
```

可選 `username` 代替 email。啟用 MFA 時加 `"mfa_token":"123456"`（TOTP 或備用碼）。回：

```json
{"status":"success","token":"<jwt>","expires_at":"...","user":{"id":"...","name":"ada","email":"ada@example.com"}}
```

連續錯密碼 5 次會鎖 15 分鐘（`err_code`: `account_locked`）。解鎖：`POST /auth/unlock-account` `{"email":"..."}` 寄信，再用 `{"email":"...","token":"..."}`。

Five failed logins lock the account for 15 minutes (`account_locked`). Unlock via email token.

**登出 / Logout** `GET /auth/logout`（Bearer 或 cookie）

**重設密碼 / Reset**

1. `POST /auth/reset-password` `{"email":"ada@example.com"}`  
2. Emailer 收到 `data["token"]`、`data["reset_url"]`（`/auth/reset-password?token=...`，請自行加 origin）  
3. `POST /auth/reset-password/confirm` `{"token":"...","password":"Newpass1!"}`  

Token 15 分鐘有效。HTML 表單成功會跳去登入頁。

**OAuth** `GET /auth/oauth/:provider` → `{auth_url, state}`；回調 `GET /auth/oauth/:provider/callback?code=` 會建／綁用戶並發 session。要先 `WithOAuthProviders`。

**CAPTCHA** `POST /auth/captcha/verify` `{"response":"..."}`

### 需登入 / Authenticated

Header：`Authorization: Bearer <token>`，或瀏覽器 cookie。

| Method | Path | 說明 / Meaning |
| --- | --- | --- |
| GET | `/auth/profile` | 資料 / profile |
| PUT/PATCH | `/auth/profile` | 改 name／email |
| GET | `/auth/profile/roles` | 角色 / roles |
| POST | `/auth/change-password` | `current_password`, `new_password` |
| POST | `/auth/change-email` | 寄確認碼到新電郵 |
| POST | `/auth/change-email/confirm` | `code`, `change_email_token` |
| POST | `/auth/verify-email/send` | 寄 6 位碼 |
| POST | `/auth/verify-email/confirm` | `{"code":"123456"}` |
| POST | `/auth/refresh-token` | 過期前 15 分鐘內換新 token |
| POST | `/auth/verify-token` | 檢查 session |
| POST | `/auth/mfa/enable` | 回 `otp_secret`、`otpauth_url`（尚未開啟） |
| POST | `/auth/mfa/verify` | 確認 TOTP 後先真正開啟；或驗證備用碼 |
| POST | `/auth/mfa/disable` | 關閉 MFA |
| POST | `/auth/mfa/backup-codes` | 一次過顯示 10 個碼（hashed 入庫） |
| GET | `/auth/sessions` | 未撤銷會話 |
| DELETE | `/auth/sessions/:id` | 撤銷指定會話（只能自己嘅） |
| POST | `/auth/revoke-tokens` | 撤銷全部會話 |
| POST | `/auth/profile/oauth/link` | `provider`, `code` |
| POST | `/auth/profile/oauth/unlink` | `provider` |
| DELETE | `/auth/account` | 刪帳號 |

MFA：`enable` 只存 secret；`verify` 成功先設 `otp_enable`。之後登入必須帶 `mfa_token`。

MFA: `enable` stores the secret only; `verify` turns it on. Later logins require `mfa_token`.

### 管理員 / Admin

`Required` + `roles` 含 `admin`。

| Method | Path |
| --- | --- |
| GET | `/auth/admin/users` |
| GET | `/auth/admin/users/search?q=` |
| GET | `/auth/admin/users/:id` |
| PUT | `/auth/admin/users/:id` |
| DELETE | `/auth/admin/users/:id` |
| PUT | `/auth/admin/users/:id/roles` `{"roles":["admin"]}` |
| GET/PUT | `/auth/admin/config` |

### 錯誤碼 / Error codes

`invalid_input`、`invalid_credentials`、`invalid_password`、`user_not_found`、`user_exists`、`invalid_token`、`unauthorized`、`forbidden`、`mfa_required`、`account_locked`、`rate_limit_exceeded`、`session_creation_failed`、`internal_server_error`。

---

## 注入依賴 / Injected services

```go
type Emailer interface {
	SendEmail(to, subject, body string) error
	SendTemplateEmail(to string, data map[string]any) error
}

type CAPTCHAService interface {
	Verify(response, clientIP string) (bool, error)
}

type OAuthProvider interface {
	GetAuthURL(state string) string
	ExchangeCode(code, redirectURI string) (accessToken, refreshToken string, err error)
	GetUserInfo(accessToken string) (map[string]any, error) // 建議含 email、id／sub、name
}

type RateLimiter interface {
	Allow(clientID string, limit int, window time.Duration) (bool, error)
}

type EventNotifier interface {
	Notify(userID, eventType, message string) error
}

type Logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
	Warnf(format string, args ...any)
}
```

`SendTemplateEmail` 常見 `data`：

| 流程 / Flow | 主要欄位 / Keys |
| --- | --- |
| 重設密碼 | `token`, `reset_url`, `duration_minute`, `user` |
| 改電郵 | `code`, `change_email_token` |
| 驗證電郵 | `code` |
| 解鎖 | `token`, `purpose`=`unlock` |

OAuth `GetUserInfo` 會讀 `email`、`id`／`sub`／`user_id`、`name`／`login`。

限流預設：登入 5 次／分鐘、註冊同重設 3 次／分鐘（按 IP）。

Default limits: login 5/min, register and reset 3/min (by IP).

```go
e2auth.Mount(eng, app,
	e2auth.WithEmailer(myMailer),
	e2auth.WithCAPTCHAService(myCaptcha),
	e2auth.WithOAuthProviders(e2auth.OAuthProviders{"google": myGoogle}),
	e2auth.WithRateLimiter(myLimiter),
	e2auth.WithEventNotifier(myEvents),
	e2auth.WithLogger(logrus.StandardLogger()),
)
```

`*logrus.Logger` 已符合 `Logger`。

---

## 選項一覽 / Options

| 選項 / Option | 說明 / Meaning |
| --- | --- |
| `WithSecretKey([]byte)` | JWT 密鑰；`Mount` 可從 `secret_key` 推導 |
| `WithAppName(string)` | HTML 頁標題 |
| `WithEmailer` | 寄信 |
| `WithCAPTCHAService` | CAPTCHA |
| `WithOAuthProviders` | OAuth |
| `WithRateLimiter` | 限流 |
| `WithLogger` | 日誌；`Mount` 預設 logrus |
| `WithEventNotifier` | 登入等事件 |
| `WithTableSchema` | Postgres schema 名（預設 `e2auth`，建 schema；表名仍係模型 `TableName`） |
| `WithDisablePages` | 唔掛 HTML 頁 |
| `WithGinTemplates` | 用 gin `c.HTML` 渲染 `login.html` 等 |
| `WithDB` | 覆寫 `*gorm.DB`（少用） |

後傳入嘅 option 會蓋過 `Mount` 自動推導嘅 secret／logger。

Later options override secrets/logger inferred by `Mount`.

---

## 你唔使做嘅嘢 / What you should not rebuild

- 用戶／會話／重設 token 表  
- 登入 JWT、bcrypt、TOTP、鎖戶  
- `/auth/login`、`/register`、`/reset-password` 同對應 HTML  

業務（訂單、內容）用 `CurrentUser(c).Id` 關聯。帳號生命週期全部行 `/auth/*`。
