# e2struct

遞迴處理 struct：trim 字串欄位，並把 nil 的 struct 指標初始化。

Recursively trim string fields and initialize nil struct pointers.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2struct
```

## 用法 / Usage

```go
import "github.com/e2u/e2util/e2struct"

type In struct {
    Name  string
    Child *In
}
v := &In{Name: "  ada  "}
e2struct.PrepareStruct(v)
// v.Name == "ada", v.Child != nil
```
