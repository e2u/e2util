# e2unit

位元組大小的人讀格式與解析（SI 與 IEC）。

Human-readable byte sizes and parsing (SI and IEC).

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2unit
```

## 功能 / Features

- `Bytes`：SI（kB、MB…）
- `IBytes`：IEC（KiB、MiB…）
- `ParseBytes`：`"42MB"`、`"42mib"`

## 用法 / Usage

```go
import "github.com/e2u/e2util/e2unit"

s := e2unit.Bytes(82854982)   // "83 MB"
s = e2unit.IBytes(82854982)   // "79 MiB"
n, err := e2unit.ParseBytes("42MB")
```
