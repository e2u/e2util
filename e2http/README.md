# e2http

鏈式 HTTP 客戶端：設 URL、標頭、逾時、Basic/Bearer、代理、把回應解成 JSON。

Fluent HTTP client: URL, headers, timeout, Basic/Bearer auth, proxy, and JSON unmarshalling.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2http
```

## 功能 / Features

- **Builder**：`URL`、`Method`、`AddHeader`、`UserAgent`、`ConnectTimeout`
- **認證 / Auth**：`SetBasicAuth`、`SetBearerAuth`
- **請求體 / Body**：`PostJSON`、`PostForm`、`PostRaw`、`PostMultipart`
- **除錯 / Debug**：`DumpRequest`、`DumpResponse`
- **MIME**：`GetContentType(filename)`

## 用法 / Usage

```go
import (
    "context"
    "github.com/e2u/e2util/e2http"
)

type Payload struct {
    ID int `json:"id"`
}

var out Payload
h := e2http.Builder(context.Background()).
    URL("https://example.com/api").
    SetBearerAuth("token").
    ToJSON(&out).
    Do()
if errs := h.Errors(); len(errs) > 0 {
    log.Fatal(errs)
}
fmt.Println(h.StatusCode(), out)
```
