# e2regexp

用命名捕捉組把 `FindStringSubmatch` 結果收成 `e2map.Map`。

Named capture groups from `FindStringSubmatch` into an `e2map.Map`.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2regexp
```

## 用法 / Usage

```go
import (
    "regexp"
    "github.com/e2u/e2util/e2regexp"
)

re := regexp.MustCompile(`(?P<area>\d{3})-(?P<line>\d{4})`)
m, ok := e2regexp.NamedFindStringSubmatch("202-0147", re)
```
