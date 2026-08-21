# e2rest

Gorm 上的薄 REST 風格 CRUD 輔助。`Create` / `Update` 已接好；其餘函式目前為佔位。

Thin Gorm REST-style CRUD helpers. `Create` / `Update` are implemented; other functions are stubs.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2rest
```

## 用法 / Usage

```go
import "github.com/e2u/e2util/e2rest"

user, err := e2rest.Create(db, &User{Name: "ada"})
user, err = e2rest.Update(db, user)
```
