# e2test

測試輔助：Gin HTTP 呼叫、隨機英文詞、列印分隔線。僅供測試使用。

Test helpers: Gin HTTP calls, random English words, and separator lines. For tests only.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2test
```

## 功能 / Features

- **Gin**：`Get`、`PostJSON`、`PutJSON`、`Any`
- **隨機文字 / Random text**：`RandomWord`、`RandomWords`、`RandomPhrase`
- **輸出 / Output**：`Line` 列印重複字元

## 用法 / Usage

```go
import "github.com/e2u/e2util/e2test"

w := e2test.Get(engine, "/ping", nil, nil)
phrase := e2test.RandomPhrase(3, 6)
```
