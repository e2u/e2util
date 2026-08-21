# e2net

網路工具的命名空間。代理檢測見子套件 [`proxychecker`](proxychecker)。

Namespace for network helpers. Proxy checks live in [`proxychecker`](proxychecker).

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2net/proxychecker
```

## 用法 / Usage

```go
import "github.com/e2u/e2util/e2net/proxychecker"

resp := proxychecker.CheckProxy(ctx, "socks5://127.0.0.1:9050",
    proxychecker.DefaultRequest("https://example.com/"))
```
