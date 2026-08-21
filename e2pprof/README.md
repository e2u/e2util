# e2pprof

在 `127.0.0.1` 隨機埠啟動 `net/http/pprof`，埠號寫進 log。

Start `net/http/pprof` on a random `127.0.0.1` port; the port is logged.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2pprof
```

## 用法 / Usage

```go
import (
    "context"
    "github.com/e2u/e2util/e2pprof"
)

e2pprof.Init(context.Background())
// 日誌會出現 http://127.0.0.1:<port>/debug/pprof
```
