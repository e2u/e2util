# e2env Documentation

## 項目概覽 / Project Overview

`e2env` 是 `e2util` 工具庫中的一個子包，用於從環境變量或命令行參數中獲取配置值。它支持字符串、布林值和整數類型的參數，優先從環境變量中取值，若無則從命令行參數取值，並提供默認值。此包適用於需要靈活配置管理的應用場景，例如服務啟動參數或環境設置。

`e2env` is a sub-package of the `e2util` library, designed to retrieve configuration values from environment variables or command-line flags. It supports string, boolean, and integer parameter types, prioritizing environment variables over command-line flags, with fallback to default values. This package is suitable for applications requiring flexible configuration management, such as service startup parameters or environment settings.

---

## 使用方法 / Usage

### 1. 獲取字符串參數 / Retrieving a String Parameter

Use `EnvStringVar` to retrieve a string parameter from environment variables or command-line flags.

```go
package main

import (
"flag"
"fmt"
"github.com/e2u/e2util/e2env"
)

func main() {
var param string
// Retrieve a string parameter (e.g., PARAM_NAME env var or --param-name flag)
e2env.EnvStringVar(&param, "param-name", "default", "Parameter name for the application")
flag.Parse()
fmt.Println("Parameter:", param)
}
```

### 2. 獲取布林值參數 / Retrieving a Boolean Parameter

Use `EnvBoolVar` to retrieve a boolean parameter from environment variables or command-line flags.

```go
package main

import (
"flag"
"fmt"
"github.com/e2u/e2util/e2env"
)

func main() {
var debug bool
// Retrieve a boolean parameter (e.g., DEBUG env var or --debug flag)
e2env.EnvBoolVar(&debug, "debug", false, "Enable debug mode")
flag.Parse()
fmt.Println("Debug mode:", debug)
}
```

### 3. 獲取整數參數 / Retrieving an Integer Parameter

Use `EnvIntVar` to retrieve an integer parameter from environment variables or command-line flags.

```go
package main

import (
"flag"
"fmt"
"github.com/e2u/e2util/e2env"
)

func main() {
var port int
// Retrieve an integer parameter (e.g., PORT env var or --port flag)
e2env.EnvIntVar(&port, "port", 8080, "Port number for the server")
flag.Parse()
fmt.Println("Port:", port)
}
```

---

## 安裝步驟 / Installation Steps

1. **確保 Go 環境**
確認已安裝 Go（建議使用 1.16 或以上版本），並設置好 `GOPATH`。
2. **下載項目**
在終端運行以下命令：
```bash
go get -u github.com/e2u/e2util/e2env
```
3. **驗證安裝**
在代碼中導入 `github.com/e2u/e2util/e2env`，運行 `go build` 或 `go run`，若無錯誤則安裝成功。

1. **Ensure Go Environment**
Confirm Go (version 1.16 or higher recommended) is installed and `GOPATH` is set.
2. **Download the Package**
Run the following command in your terminal:
```bash
go get -u github.com/e2u/e2util/e2env
```
3. **Verify Installation**
Import `github.com/e2u/e2util/e2env` in your code and run `go build` or `go run`. Success indicates proper installation.

---

## 功能描述 / Features

- **環境變量優先**：優先從環境變量中獲取參數值，若無則從命令行參數取值。
- **多類型支持**：支持字符串（`EnvStringVar`）、布林值（`EnvBoolVar`）和整數（`EnvIntVar`）參數。
- **鍵名轉換**：命令行參數中的 `-` 自動轉換為環境變量中的 `_`（例如 `param-name` 對應 `PARAM_NAME`）。
- **默認值支持**：提供默認值，當環境變量和命令行參數均未設置時使用。
- **使用簡單**：與 `flag` 包集成，需在程序中調用 `flag.Parse()` 完成參數解析。

- **Environment Variable Priority**: Prioritizes environment variables for parameter values, falling back to command-line flags if not set.
- **Multi-Type Support**: Supports string (`EnvStringVar`), boolean (`EnvBoolVar`), and integer (`EnvIntVar`) parameters.
- **Key Conversion**: Converts `-` in command-line flags to `_` in environment variables (e.g., `param-name` maps to `PARAM_NAME`).
- **Default Value Support**: Provides default values when neither environment variables nor command-line flags are set.
- **Simple Usage**: Integrates with the `flag` package, requiring a call to `flag.Parse()` to parse parameters.

---
```
