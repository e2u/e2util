# e2context

檢查 `context` 是否仍有效，然後呼叫 cancel 並確認已取消。

Check whether a context is still active, then cancel it and confirm.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2context
```

## 用法 / Usage

```go
import (
    "context"
    "github.com/e2u/e2util/e2context"
)

ctx, cancel := context.WithCancel(context.Background())
e2context.CheckAndCancelContext(ctx, cancel)
```
