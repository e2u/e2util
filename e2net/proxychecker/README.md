# proxychecker

透過指定 HTTP/SOCKS 代理去請求目標 URL，並用正則檢查回應是否符合預期。

Fetch a target URL through an HTTP/SOCKS proxy and match the response with regexes.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2net/proxychecker
```

## 功能 / Features

- **DefaultRequest**：預設 UA、逾時、期望 `HTTP/x.x 200 OK`
- **鏈式設定 / Builders**：`WithExpect`、`WithNotExpect`、`WithTimeout`、`WithHttpMethod`
- **CheckProxy**：回傳狀態碼、耗時、是否命中 expect / not-expect

真實代理測試需設定環境變數 `E2UTIL_PROXY`。

Live proxy tests require `E2UTIL_PROXY`.

## 用法 / Usage

```go
import (
    "context"
    "github.com/e2u/e2util/e2net/proxychecker"
)

resp := proxychecker.CheckProxy(context.Background(),
    "socks5://127.0.0.1:9050",
    proxychecker.DefaultRequest("https://example.com/"),
)
if resp.Error != nil {
    log.Fatal(resp.Error)
}
```
