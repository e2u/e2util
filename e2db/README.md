# e2db Documentation

## 項目概覽 / Project Overview

`e2db` 是 `e2util` 工具庫中的一個子包，提供了一個基於 `gorm` 的數據庫操作框架。它支持多種數據庫（如 MySQL、PostgreSQL、SQLite），實現讀寫分離、自動遷移、JSONB 字段存儲、軟刪除模型和日誌記錄。此包適用於需要高效數據庫操作的應用場景，例如數據存儲、查詢和模型管理。

`e2db` is a sub-package of the `e2util` library, providing a database operation framework based on `gorm`. It supports multiple databases (e.g., MySQL, PostgreSQL, SQLite), implements read-write separation, auto-migration, JSONB field storage, soft delete models, and logging. This package is suitable for applications requiring efficient database operations, such as data storage, querying, and model management.

---

## 使用方法 / Usage

### 1. 初始化數據庫連接 / Initializing Database Connection

Use the `New` function to create a database connection with read-write separation.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2db"
)

func main() {
// Configure database connection
cfg := &e2db.Config{
Writer:  "host=localhost port=5432 user=postgres password=pass dbname=mydb sslmode=disable",
Readers: []string{"host=localhost port=5432 user=postgres password=pass dbname=mydb sslmode=disable"},
Driver:  "postgres",
}
conn, err := e2db.New(cfg)
if err != nil {
fmt.Println("Failed to initialize database:", err)
return
}
fmt.Println("Database connection initialized")
}
```

### 2. 計數查詢 / Counting Records

Use the `Count` method to count records matching a query.

```go
package main

import (
"context"
"fmt"
"github.com/e2u/e2util/e2db"
)

type User struct {
ID   uint
Name string
}

func main() {
cfg := &e2db.Config{Writer: "file::memory:?cache=shared", Driver: "sqlite"}
conn, _ := e2db.New(cfg)
ctx := context.Background()

// Count users with a specific name
count, err := conn.Count(ctx, &User{}, true, "name = ?", "Alice")
if err != nil {
fmt.Println("Count failed:", err)
return
}
fmt.Println("Count:", count.Int64)
}
```

### 3. 保存並預加載模型 / Saving and Preloading a Model

Use `DBHandler` to save a model and preload its associations.

```go
package main

import (
"context"
"fmt"
"github.com/e2u/e2util/e2db"
)

type User struct {
ID      uint
Name    string
Profile Profile
}

type Profile struct {
ID     uint
UserID uint
Bio    string
}

func main() {
cfg := &e2db.Config{Writer: "file::memory:?cache=shared", Driver: "sqlite"}
conn, _ := e2db.New(cfg)
ctx := context.Background()

// Create a DBHandler for User
handler := e2db.NewDBHandler[User](conn)
user := User{Name: "Bob"}

// Save and preload the user with associations
savedUser, err := handler.SaveAndPreload(ctx, user)
if err != nil {
fmt.Println("Save and preload failed:", err)
return
}
fmt.Println("Saved user:", savedUser.Name)
}
```

### 4. 自動遷移模型 / Auto-Migrating Models

Use `AutoMigrate` to automatically create or update database tables based on models.

```go
package main

import (
"context"
"fmt"
"github.com/e2u/e2util/e2db"
)

type Product struct {
ID   uint
Name string
}

func main() {
cfg := &e2db.Config{Writer: "file::memory:?cache=shared", Driver: "sqlite"}
conn, _ := e2db.New(cfg)
ctx := context.Background()

// Auto-migrate the Product model
err := conn.AutoMigrate(ctx, &Product{})
if err != nil {
fmt.Println("Auto-migrate failed:", err)
return
}
fmt.Println("Table migrated successfully")
}
```

### 5. 創建 PostgreSQL 模式 / Creating PostgreSQL Schemas

Use `CreateSchema` to create PostgreSQL schemas (only supported for PostgreSQL).

```go
package main

import (
"context"
"fmt"
"github.com/e2u/e2util/e2db"
)

func main() {
cfg := &e2db.Config{
Writer: "host=localhost port=5432 user=postgres password=pass dbname=mydb sslmode=disable",
Driver: "postgres",
}
conn, _ := e2db.New(cfg)
ctx := context.Background()

// Create a new schema
err := conn.CreateSchema(ctx, "public")
if err != nil {
fmt.Println("Create schema failed:", err)
return
}
fmt.Println("Schema created successfully")
}
```

### 6. 檢查記錄是否存在 / Checking if a Record Exists

Use `Exists` to check if a record exists in the database.

```go
package main

import (
"context"
"fmt"
"github.com/e2u/e2util/e2db"
)

type User struct {
ID   uint
Name string
}

func main() {
cfg := &e2db.Config{Writer: "file::memory:?cache=shared", Driver: "sqlite"}
conn, _ := e2db.New(cfg)
ctx := context.Background()

// Check if a user exists
exists := conn.Exists(&User{}, "name = ?", true, "Alice")
if exists.Error != nil {
fmt.Println("Check failed:", exists.Error)
return
}
fmt.Println("Exists:", exists.Bool)
}
```

### 7. 使用 JSONB 字段存儲數據 / Storing Data with JSONB Fields

Use `JSONBArray` and `JSONBMap` to store JSONB data in the database.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2db"
)

type Product struct {
ID         uint
Name       string
Attributes e2db.JSONBMap
}

func main() {
cfg := &e2db.Config{Writer: "file::memory:?cache=shared", Driver: "sqlite"}
conn, _ := e2db.New(cfg)

// Create a product with JSONB attributes
product := Product{
Name: "Laptop",
Attributes: e2db.JSONBMap{Data: map[string]any{
"color": "silver",
"price": 999.99,
}},
}
err := conn.RW().Create(&product).Error
if err != nil {
fmt.Println("Create failed:", err)
return
}
fmt.Println("Product created with JSONB attributes")
}
```

---

## 安裝步驟 / Installation Steps

1. **確保 Go 環境**
確認已安裝 Go（建議使用 1.16 或以上版本），並設置好 `GOPATH`。
2. **下載項目**
在終端運行以下命令：
```bash
go get -u github.com/e2u/e2util/e2db
```
3. **驗證安裝**
在代碼中導入 `github.com/e2u/e2util/e2db`，運行 `go build` 或 `go run`，若無錯誤則安裝成功。

1. **Ensure Go Environment**
Confirm Go (version 1.16 or higher recommended) is installed and `GOPATH` is set.
2. **Download the Package**
Run the following command in your terminal:
```bash
go get -u github.com/e2u/e2util/e2db
```
3. **Verify Installation**
Import `github.com/e2u/e2util/e2db` in your code and run `go build` or `go run`. Success indicates proper installation.

---

## 功能描述 / Features

- **讀寫分離**：支持主寫從讀（RW/RO）模式，通過 `RO` 和 `RW` 方法選擇數據庫連接。
- **多數據庫支持**：支持 MySQL、PostgreSQL 和 SQLite，自動檢測和配置數據庫類型。
- **自動遷移**：`AutoMigrate` 自動創建或更新數據庫表結構。
- **JSONB 存儲**：`JSONBArray` 和 `JSONBMap` 支持存儲 JSONB 數據，適用於 PostgreSQL。
- **軟刪除模型**：`ModelWithSoftDelete` 提供軟刪除功能，自動記錄刪除時間。
- **計數和存在檢查**：`Count` 和 `Exists` 方法支持高效查詢記錄數量和存在性。
- **日誌記錄**：集成 `gorm` 日誌，支持慢查詢記錄和自定義日誌配置。
- **模式管理**：`CreateSchema` 支持在 PostgreSQL 中創建模式。
- **錯誤處理**：提供詳細錯誤信息，支持自動創建數據庫和初始化 SQL 執行。

- **Read-Write Separation**: Supports read-write separation (RW/RO) mode, selecting database connections via `RO` and `RW` methods.
- **Multi-Database Support**: Supports MySQL, PostgreSQL, and SQLite, with automatic detection and configuration.
- **Auto-Migration**: `AutoMigrate` automatically creates or updates database table structures.
- **JSONB Storage**: `JSONBArray` and `JSONBMap` support storing JSONB data, suitable for PostgreSQL.
- **Soft Delete Model**: `ModelWithSoftDelete` provides soft delete functionality, automatically recording deletion time.
- **Counting and Existence Check**: `Count` and `Exists` methods enable efficient record counting and existence checks.
- **Logging**: Integrates `gorm` logging, supporting slow query logging and custom log configuration.
- **Schema Management**: `CreateSchema` supports creating schemas in PostgreSQL.
- **Error Handling**: Provides detailed error information, supporting automatic database creation and initialization SQL execution.

---
```
