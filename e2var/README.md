# e2var

指標、零值預設、條件取值。不是讀環境變數（請用 `e2env`）。

Pointers, zero-value defaults, and conditional values. Not env vars (use `e2env`).

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2var
```

## 功能 / Features

- **指標 / Pointers**：`P`、`MustStringValue`、`NeverNullPoint`
- **預設 / Defaults**：`NeverNull`、`ValueOrDefault`、`ExpectOrDefault`
- **條件 / Branch**：`IfElse`、`TrueThen`、`NullThen`、`NotNullThen`

## 用法 / Usage

```go
import "github.com/e2u/e2util/e2var"

n := e2var.P(3)
s := e2var.NeverNull("", "fallback")
v := e2var.TrueThen(ok, "yes", "no")
```
