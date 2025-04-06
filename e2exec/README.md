---

# e2exec Documentation

## 項目概覽 / Project Overview

`e2exec` 是 `e2util` 工具庫中的一個子包，提供了一組用於錯誤處理和條件執行的工具函數。它包含兩大功能模塊：
1. **錯誤處理**：提供多種函數來簡化錯誤檢查和日誌記錄，支持泛型返回值、靜默錯誤處理以及資源關閉操作，適用於需要健壯錯誤管理的場景。
2. **條件執行**：支持基於布林值或空值檢查的條件函數執行，提供靈活的控制流處理，適用於動態邏輯分支或默認行為設置。
此包依賴 Go 的標準庫（如 `runtime` 和 `reflect`）以及 `github.com/sirupsen/logrus` 日誌庫，確保高效性和可維護性。

`e2exec` is a sub-package of the `e2util` library, offering a set of utility functions for error handling and conditional execution. It includes two main functional modules:
1. **Error Handling**: Provides functions to simplify error checking and logging, supporting generic return values, silent error handling, and resource closure, suitable for scenarios requiring robust error management.
2. **Conditional Execution**: Supports conditional function execution based on boolean or nil checks, offering flexible control flow handling, suitable for dynamic logic branching or default behavior setup.
This package relies on Go’s standard library (e.g., `runtime` and `reflect`) and the `github.com/sirupsen/logrus` logging library, ensuring efficiency and maintainability.

---

## 使用方法 / Usage

### 1. 檢查單個返回值並記錄錯誤 / Checking a Single Return Value and Logging Errors

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2exec"
)

func main() {
// 模擬一個可能失敗的函數 / Simulate a function that might fail
result := e2exec.Must(someFunction())
fmt.Println("結果 / Result:", result)
}

func someFunction() (string, error) {
return "成功 / Success", nil
}
```

### 2. 檢查雙返回值並記錄錯誤 / Checking Dual Return Values and Logging Errors

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2exec"
)

func main() {
// 處理雙返回值 / Handle dual return values
v1, v2 := e2exec.Must2(someDualFunction())
fmt.Println("值1 / Value 1:", v1, "值2 / Value 2:", v2)
}

func someDualFunction() (int, string, error) {
return 42, "測試 / Test", nil
}
```

### 3. 關閉資源並記錄錯誤 / Closing Resources and Logging Errors

```go
package main

import (
"os"
"github.com/e2u/e2util/e2exec"
)

func main() {
// 打開文件並確保關閉 / Open a file and ensure closure
file, _ := os.Open("test.txt")
defer e2exec.MustClose(file)
}
```

### 4. 靜默處理錯誤 / Silently Handling Errors

```go
package main

import (
"errors"
"github.com/e2u/e2util/e2exec"
)

func main() {
// 靜默記錄錯誤 / Silently log an error
err := errors.New("某個錯誤 / Some error")
e2exec.SilentError("操作失敗 / Operation failed", err)
}
```

### 5. 僅返回錯誤 / Returning Only Errors

```go
package main

import (
"errors"
"fmt"
"github.com/e2u/e2util/e2exec"
)

func main() {
// 提取並返回錯誤 / Extract and return an error
err := e2exec.OnlyError(nil, errors.New("測試錯誤 / Test error"))
if err != nil {
fmt.Println("錯誤 / Error:", err)
}
}
```

### 6. 基於布林值選擇函數 / Selecting Functions Based on Boolean

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2exec"
)

func main() {
// 根據布林值選擇函數 / Select function based on boolean
fn := e2exec.TrueThen(true, func() { fmt.Println("真 / True") }, func() { fmt.Println("假 / False") })
fn()
}
```

### 7. 基於布林值執行並返回值 / Executing and Returning Based on Boolean

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2exec"
)

func main() {
// 根據布林值執行並返回值 / Execute and return based on boolean
result := e2exec.TrueThenExec(false, func() any { return "真 / True" }, func() any { return "假 / False" })
fmt.Println("結果 / Result:", result)
}
```

### 8. 檢查非空值執行函數 / Executing Functions Based on Non-Null Check

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2exec"
)

func main() {
// 檢查值是否非空 / Check if value is non-null
var ptr *string
e2exec.NotNullThenFunc(ptr, func() { fmt.Println("非空 / Not Null") }, func() { fmt.Println("空 / Null") })
}
```

---

## 安裝步驟 / Installation Steps

1. **確保 Go 環境**
確認已安裝 Go（建議使用 1.16 或以上版本），並設置好 `GOPATH`。
Ensure Go (version 1.16 or higher recommended) is installed and `GOPATH` is set.

2. **下載項目**
在終端運行以下命令 / Run the following command in your terminal:
```bash
go get -u github.com/e2u/e2util/e2exec
```

3. **驗證安裝**
在代碼中導入 `github.com/e2u/e2util/e2exec`，運行 `go build` 或 `go run`，若無錯誤則安裝成功。
Import `github.com/e2u/e2util/e2exec` in your code and run `go build` or `go run`. Success indicates proper installation.

---

## 功能描述 / Features

- **單返回值檢查**：`Must` 檢查單個返回值並記錄錯誤，支持泛型類型。
**Single Return Value Check**: `Must` checks a single return value and logs errors, supporting generic types.

- **雙返回值檢查**：`Must2` 檢查兩個返回值並記錄錯誤，支持泛型類型。
**Dual Return Value Check**: `Must2` checks two return values and logs errors, supporting generic types.

- **資源關閉**：`MustClose` 確保資源（如文件或連接）正確關閉並記錄錯誤。
**Resource Closure**: `MustClose` ensures resources (e.g., files or connections) are closed properly and logs errors.

- **靜默錯誤處理**：`SilentError` 記錄錯誤並提供調用棧信息，適用於不需要拋出錯誤的場景。
**Silent Error Handling**: `SilentError` logs errors with call stack details, suitable for scenarios where errors don’t need to be thrown.

- **僅返回錯誤**：`OnlyError` 提取並返回錯誤，同時記錄日誌，簡化錯誤處理流程。
**Error-Only Return**: `OnlyError` extracts and returns errors while logging, simplifying error handling workflows.

- **布林條件執行**：`TrueThen` 和 `TrueThenExec` 根據布林值選擇並執行函數，支持返回值。
**Boolean Conditional Execution**: `TrueThen` and `TrueThenExec` select and execute functions based on boolean values, supporting return values.

- **空值檢查執行**：`NotNullThenFunc` 和 `NullThenFunc` 根據值的空狀態執行不同函數，支持反射檢查。
**Nil Check Execution**: `NotNullThenFunc` and `NullThenFunc` execute functions based on nil status, using reflection for robustness.

- **錯誤處理**：所有函數均針對無效輸入（如空參數或非錯誤類型）進行健壯處理，並通過 `logrus` 記錄詳細日誌。
**Error Handling**: All functions handle invalid inputs (e.g., empty args or non-error types) robustly, logging detailed info via `logrus`.

---
