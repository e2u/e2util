# e2json

JSON 輔助函數，以及能接受數字或字串的寬鬆型別（Flex*）。

JSON helpers plus flexible types that accept numbers or strings.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2json
```

## 功能 / Features

- **序列化 / Marshal**：`MustToJSONByte`、`MustToJSONString`、`MustIndentJSONString`
- **反序列化 / Unmarshal**：`MustFromJSONByte`、`MustFromJSONString`、`MustFromReader`
- **日期 / Date**：`NowDate`、`ParseDate`（可用 `e2json.DateFormat` 改格式）
- **寬鬆欄位 / Flex types**：`FlexInt64`、`FlexFloat64`、`FlexBool`、`FlexDecimal`

## 用法 / Usage

```go
import (
    "encoding/json"
    "github.com/e2u/e2util/e2json"
)

s := e2json.MustToJSONString(map[string]any{"ok": true})
var m map[string]any
_ = e2json.MustFromJSONString(s, &m)

type Row struct {
    N e2json.FlexInt64 `json:"n"`
}
var row Row
_ = json.Unmarshal([]byte(`{"n":"42"}`), &row)
n, ok := row.N.Value() // 42, true
```
