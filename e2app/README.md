# e2app

載入 TOML／環境變數／命令列，組出帶 DB、快取、HTTP、日誌的應用 `Context`。

Loads TOML, env, and flags into an application `Context` with DB, cache, HTTP, and logger.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2app
```

## 功能 / Features

- **設定來源 / Config sources**：`--config` 檔、embed.FS（如 `dev.toml`）、`.` / `etc` / `conf` 等路徑
- **組件 / Wiring**：`e2db`、`e2cache`、`e2http.Config`、`e2logrus`
- **AppConfig**：`Get` / `GetString` / `GetInt` / `GetStringSlice` / `GetBytesFromBase64`

Postgres 不可用時，依賴資料庫的測試會 skip。

Tests that need Postgres skip when it is unavailable.

## 用法 / Usage

```go
import (
    "context"
    "embed"
    "github.com/e2u/e2util/e2app"
)

//go:embed *.toml
var cfgFS embed.FS

func main() {
    ctx := e2app.New(context.Background(), cfgFS)
    _ = ctx.DB
    _ = ctx.Cache
    _ = ctx.App.GetString("secret_key")
}
```

認證模組見 [`e2auth.Mount`](../e2auth)：把 `e2app` 嘅 DB 同 `secret_key` 交給 `e2gin` 引擎。

For login/register, see [`e2auth.Mount`](../e2auth) — it takes this `Context` and an `e2gin` engine.
