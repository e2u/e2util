# e2gintest

用 `httptest` 註冊並呼叫單個 Gin handler，方便單元測試。

Register and invoke a single Gin handler via `httptest`.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2gintest
```

## 用法 / Usage

```go
import (
    "net/http"
    "github.com/e2u/e2util/e2gintest"
    "github.com/gin-gonic/gin"
)

g := e2gintest.New()
w := g.Run(&e2gintest.Request{
    ReqUri: "/ping",
    Method: http.MethodGet,
    Handlers: []gin.HandlerFunc{
        func(c *gin.Context) { c.String(200, "pong") },
    },
})
```
