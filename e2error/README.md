# e2error Documentation

## 項目概覽 / Project Overview

`e2error` 是 `e2util` 工具庫中的一個子包，提供了一組用於錯誤處理的工具函數。它包含預定義的錯誤生成器（如配置錯誤、非法參數等），並提供檢查和記錄多個錯誤的實用方法。此包適用於需要統一錯誤管理和日誌記錄的應用場景，例如服務端錯誤處理或批量操作。

`e2error` is a sub-package of the `e2util` library, providing a set of utility functions for error handling. It includes predefined error generators (e.g., configuration errors, illegal parameters) and offers methods for checking and logging multiple errors. This package is suitable for scenarios requiring unified error management and logging, such as server-side error handling or batch operations.

---

## 使用方法 / Usage

### 1. 使用預定義錯誤生成器 / Using Predefined Error Generators

Use predefined error functions to create specific error types.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2error"
)

func main() {
// Create a configuration error
err := e2error.ErrConfigureError("invalid database DSN")
fmt.Println("Error:", err) // Error: configuration error invalid database DSN
}
```

### 2. 檢查多個錯誤（停止模式） / Checking Multiple Errors (Stop Mode)

Use `CheckErrors` with `stop=true` to return the first non-nil error encountered.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2error"
)

func main() {
// Check multiple errors, stopping at the first non-nil error
err := e2error.CheckErrors(true,
nil,
e2error.ErrIllegalParameter("user_id"),
e2error.ErrEmptyValue("email"),
)
fmt.Println("First error:", err) // First error: illegal parameter user_id
}
```

### 3. 檢查多個錯誤（非停止模式） / Checking Multiple Errors (Non-Stop Mode)

Use `CheckErrors` with `stop=false` to return the last non-nil error.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2error"
)

func main() {
// Check multiple errors, returning the last non-nil error
err := e2error.CheckErrors(false,
nil,
e2error.ErrIllegalParameter("user_id"),
e2error.ErrEmptyValue("email"),
)
fmt.Println("Last error:", err) // Last error: value email empty
}
```

### 4. 靜默檢查錯誤函數 / Silently Checking Error Functions

Use `SilentCheckFunc` to execute functions and log any errors without stopping.

```go
package main

import (
"github.com/e2u/e2util/e2error"
)

func main() {
// Define functions that may return errors
f1 := func() error { return nil }
f2 := func() error { return e2error.ErrUnknown("something went wrong") }

// Silently check and log errors from functions
e2error.SilentCheckFunc(f1, f2) // Logs: Received error: unknown error something went wrong
}
```

### 5. 靜默檢查錯誤列表 / Silently Checking Error List

Use `SilentCheckErrs` to log a list of errors without stopping.

```go
package main

import (
"github.com/e2u/e2util/e2error"
)

func main() {
// Silently check and log a list of errors
e2error.SilentCheckErrs(
nil,
e2error.ErrEmptyParameter("username"),
e2error.ErrIllegalParameter("age"),
) // Logs: Received error: parameter username can't not empty
// Logs: Received error: illegal parameter age
}
```

---

## 安裝步驟 / Installation Steps

1. **確保 Go 環境**
確認已安裝 Go（建議使用 1.16 或以上版本），並設置好 `GOPATH`。
2. **下載項目**
在終端運行以下命令：
```bash
go get -u github.com/e2u/e2util/e2error
```
3. **驗證安裝**
在代碼中導入 `github.com/e2u/e2util/e2error`，運行 `go build` 或 `go run`，若無錯誤則安裝成功。

1. **Ensure Go Environment**
Confirm Go (version 1.16 or higher recommended) is installed and `GOPATH` is set.
2. **Download the Package**
Run the following command in your terminal:
```bash
go get -u github.com/e2u/e2util/e2error
```
3. **Verify Installation**
Import `github.com/e2u/e2util/e2error` in your code and run `go build` or `go run`. Success indicates proper installation.

---

## 功能描述 / Features

- **預定義錯誤**：提供多種錯誤生成器（如 `ErrConfigureError`、`ErrIllegalParameter`），便於統一錯誤格式。
- **錯誤檢查**：`CheckErrors` 支持檢查多個錯誤，可選擇停止或返回最後一個非空錯誤。
- **靜默錯誤處理**：`SilentCheckFunc` 和 `SilentCheckErrs` 靜默記錄錯誤，適用於非關鍵操作。
- **日誌集成**：與 `logrus` 集成，自動記錄錯誤信息。
- **靈活使用**：支持動態錯誤消息，適配不同場景。

- **Predefined Errors**: Provides multiple error generators (e.g., `ErrConfigureError`, `ErrIllegalParameter`) for consistent error formatting.
- **Error Checking**: `CheckErrors` supports checking multiple errors, with options to stop or return the last non-nil error.
- **Silent Error Handling**: `SilentCheckFunc` and `SilentCheckErrs` silently log errors, suitable for non-critical operations.
- **Logging Integration**: Integrates with `logrus` for automatic error logging.
- **Flexible Usage**: Supports dynamic error messages, adaptable to various scenarios.

---
```
