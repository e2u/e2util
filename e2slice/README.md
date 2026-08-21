# e2slice

字串／數字切片常用操作：包含、合併、去重、安全取值。

Common slice helpers: contains, merge, unique, safe index.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2slice
```

## 功能 / Features

- **包含 / Contains**：`IncludeString`、`ContainInt64`、`BoolSliceInclude`、`ContainString`
- **前後綴 / Prefix-suffix**：`HasPrefix`、`HasSuffix`
- **集合 / Set-like**：`MergeStringSlice`、`UniqStringSlice`、`CompareStringSlice`
- **泛型 / Generic**：`GetDefault`、`Copy`、`HasConsecutiveNumbers`

## 用法 / Usage

```go
import "github.com/e2u/e2util/e2slice"

ok := e2slice.IncludeString([]string{"a", "b"}, "b")
uniq := e2slice.UniqStringSlice([]string{"a", "a", "b"})
v := e2slice.GetDefault([]int{1, 2}, 9, -1) // -1
```
