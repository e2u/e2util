# e2model

HTTP JSON Patch 模型，以及帶 error 的 `NullBool`。

HTTP JSON Patch model and a `NullBool` that carries an error.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2model
```

## 功能 / Features

- **HttpPatch**：`op` / `path` / `value`；`AllowOp`、`AllowPath`
- **NullBool**：`NewNullBool(b, err)`，`Valid` 在 err==nil 時為 true

## 用法 / Usage

```go
import "github.com/e2u/e2util/e2model"

p := &e2model.HttpPatch{Op: e2model.HttpPatchOpReplace, Path: "name", Value: "ada"}
if p.AllowOp([]string{e2model.HttpPatchOpReplace}) {
    // apply
}
nb := e2model.NewNullBool(true, nil)
```
