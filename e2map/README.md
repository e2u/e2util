# e2map

執行緒安全的 `map[string]any`，以及帶鎖的 Keys/Values/LoadOrStore。

Thread-safe `map[string]any` plus locked Keys/Values/LoadOrStore.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2map
```

## 功能 / Features

- **Map 方法**：`Get`、`Set`、`KeyExists`、`DefaultGet`、`DefaultString`、`DefaultInt`、`DefaultBool`、`DecodeBase64Value`
- **泛型函式 / Generic**：`Keys`、`Values`、`DefaultValue`、`LoadOrStore`、`Get[K,V]`

## 用法 / Usage

```go
import "github.com/e2u/e2util/e2map"

m := e2map.Map{"a": 1}
v, ok := m.Get("a")
s, _ := m.DefaultString("missing", "n/a")
```
