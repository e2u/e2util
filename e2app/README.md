# e2app Documentation

## 項目概覽 / Project Overview

`e2app` 是 `e2util` 工具庫中的一個子包，用於管理應用程序的配置和上下文。它提供了靈活的配置解析功能，支持從環境變數、命令行參數或配置文件（如 TOML 格式）中加載配置，並集成了數據庫（`e2db`）、緩存（`e2cache`）、HTTP（`e2http`）和日誌（`e2logrus`）等模塊。此包適用於需要集中管理和初始化應用程序配置的場景。

`e2app` is a sub-package of the `e2util` library, designed for managing application configuration and context. It provides flexible configuration parsing, supporting loading configurations from environment variables, command-line flags, or configuration files (e.g., TOML format), and integrates with database (`e2db`), cache (`e2cache`), HTTP (`e2http`), and logging (`e2logrus`) modules. This package is suitable for scenarios requiring centralized management and initialization of application configurations.

---

## 使用方法 / Usage

### 1. 初始化應用上下文 / Initializing Application Context

使用 `New` 函數初始化應用上下文，自動加載配置文件並設置環境。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2app"
)

func main() {
// 初始化應用上下文
ctx := e2app.New()
fmt.Println("環境:", ctx.Env)
fmt.Println("應用名稱:", ctx.App.Name)
}
```

Use the `New` function to initialize the application context, automatically loading configuration files and setting up the environment.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2app"
)

func main() {
// Initialize the application context
ctx := e2app.New()
fmt.Println("Environment:", ctx.Env)
fmt.Println("App Name:", ctx.App.Name)
}
```

### 2. 從應用配置中獲取字符串 / Getting a String from App Configuration

使用 `GetString` 方法從應用配置中獲取字符串值。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2app"
)

func main() {
ctx := e2app.New()
// 獲取字符串值
value := ctx.App.GetString("abc")
fmt.Println("字符串值:", value) // 例如 "ffefef"
}
```

Use the `GetString` method to retrieve a string value from the application configuration.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2app"
)

func main() {
ctx := e2app.New()
// Get a string value
value := ctx.App.GetString("abc")
fmt.Println("String value:", value) // e.g., "ffefef"
}
```

### 3. 從應用配置中獲取整數 / Getting an Integer from App Configuration

使用 `GetInt` 方法從應用配置中獲取整數值。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2app"
)

func main() {
ctx := e2app.New()
// 獲取整數值
value := ctx.App.GetInt("ccc")
fmt.Println("整數值:", value) // 例如 12345
}
```

Use the `GetInt` method to retrieve an integer value from the application configuration.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2app"
)

func main() {
ctx := e2app.New()
// Get an integer value
value := ctx.App.GetInt("ccc")
fmt.Println("Integer value:", value) // e.g., 12345
}
```

### 4. 從應用配置中獲取布林值 / Getting a Boolean from App Configuration

使用 `GetBool` 方法從應用配置中獲取布林值。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2app"
)

func main() {
ctx := e2app.New()
// 獲取布林值
debug := ctx.App.GetBool("settings.debug")
fmt.Println("布林值:", debug) // 例如 true
}
```

Use the `GetBool` method to retrieve a boolean value from the application configuration.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2app"
)

func main() {
ctx := e2app.New()
// Get a boolean value
debug := ctx.App.GetBool("settings.debug")
fmt.Println("Boolean value:", debug) // e.g., true
}
```

### 5. 從應用配置中獲取字符串切片 / Getting a String Slice from App Configuration

使用 `GetStringSlice` 方法從應用配置中獲取字符串切片。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2app"
)

func main() {
ctx := e2app.New()
// 獲取字符串切片
tags := ctx.App.GetStringSlice("tags")
fmt.Println("字符串切片:", tags) // 例如 ["golang", "viper", "config"]
}
```

Use the `GetStringSlice` method to retrieve a string slice from the application configuration.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2app"
)

func main() {
ctx := e2app.New()
// Get a string slice
tags := ctx.App.GetStringSlice("tags")
fmt.Println("String slice:", tags) // e.g., ["golang", "viper", "config"]
}
```

### 6. 從應用配置中獲取映射 / Getting a Map from App Configuration

使用 `GetStringMap` 方法從應用配置中獲取映射。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2app"
)

func main() {
ctx := e2app.New()
// 獲取映射
settings := ctx.App.GetStringMap("settings")
fmt.Println("映射:", settings) // 例如 map[debug:true timeout:30 theme:dark]
}
```

Use the `GetStringMap` method to retrieve a map from the application configuration.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2app"
)

func main() {
ctx := e2app.New()
// Get a map
settings := ctx.App.GetStringMap("settings")
fmt.Println("Map:", settings) // e.g., map[debug:true timeout:30 theme:dark]
}
```

### 7. 從 Base64 字符串解碼字節 / Decoding Bytes from Base64 String

使用 `GetBytesFromBase64` 方法從 Base64 編碼的字符串解碼字節數據。

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2app"
)

func main() {
ctx := e2app.New()
// 從 Base64 字符串解碼字節
bytes := ctx.App.GetBytesFromBase64("secret_key")
fmt.Println("解碼字節:", string(bytes)) // 例如 "secret_key"
}
```

Use the `GetBytesFromBase64` method to decode bytes from a Base64-encoded string.

```go
package main

import (
"fmt"
"github.com/e2u/e2util/e2app"
)

func main() {
ctx := e2app.New()
// Decode bytes from a Base64 string
bytes := ctx.App.GetBytesFromBase64("secret_key")
fmt.Println("Decoded bytes:", string(bytes)) // e.g., "secret_key"
}
```

---

## 安裝步驟 / Installation Steps

1. **確保 Go 環境**
確認已安裝 Go（建議使用 1.16 或以上版本），並設置好 `GOPATH`。
2. **下載項目**
在終端運行以下命令：
```bash
go get -u github.com/e2u/e2util/e2app
```
3. **驗證安裝**
在代碼中導入 `github.com/e2u/e2util/e2app`，運行 `go build` 或 `go run`，若無錯誤則安裝成功。

1. **Ensure Go Environment**
Confirm Go (version 1.16 or higher recommended) is installed and `GOPATH` is set.
2. **Download the Package**
Run the following command in your terminal:
```bash
go get -u github.com/e2u/e2util/e2app
```
3. **Verify Installation**
Import `github.com/e2u/e2util/e2app` in your code and run `go build` or `go run`. Success indicates proper installation.

---

## 功能描述 / Features

- **應用上下文管理**：`New` 函數初始化應用上下文，支持從環境變數、命令行參數或配置文件（如 TOML）中加載配置，並集成數據庫、緩存、HTTP 和日誌模塊。
- **靈活配置存取**：`AppConfig` 結構提供多種方法（如 `GetString`、`GetInt`、`GetBool`、`GetStringSlice`、`GetStringMap`）從配置中獲取不同類型的值。
- **Base64 解碼**：`GetBytesFromBase64` 方法支持從 Base64 編碼字符串解碼字節數據。
- **配置文件支持**：支持多種數據庫配置（如 PostgreSQL、SQLite）、緩存配置（如 Redis）、HTTP 配置和日誌配置，通過 TOML 文件靈活設置。
- **環境適配**：支持通過命令行參數（如 `--env`）或環境變數（如 `ENV`）指定運行環境（如 `dev`、`test`、`prod`）。

- **Application Context Management**: The `New` function initializes the application context, supporting loading configurations from environment variables, command-line flags, or configuration files (e.g., TOML), and integrates with database, cache, HTTP, and logging modules.
- **Flexible Configuration Access**: The `AppConfig` struct provides various methods (e.g., `GetString`, `GetInt`, `GetBool`, `GetStringSlice`, `GetStringMap`) to retrieve different types of values from the configuration.
- **Base64 Decoding**: The `GetBytesFromBase64` method supports decoding byte data from Base64-encoded strings.
- **Configuration File Support**: Supports configurations for multiple databases (e.g., PostgreSQL, SQLite), caches (e.g., Redis), HTTP, and logging, flexibly set via TOML files.
- **Environment Adaptation**: Supports specifying the runtime environment (e.g., `dev`, `test`, `prod`) via command-line flags (e.g., `--env`) or environment variables (e.g., `ENV`).

---

## 配置文件示例 / Configuration File Example

以下是配置文件 `dev.toml` 的示例，展示如何設置應用配置、數據庫、緩存、HTTP 和日誌模塊。

```toml
[app]
name = "App Name"
tags = ["golang", "viper", "config"]
abc = "ffefef"
ccc = 12345
secret_key = "c2VjcmV0X2tleQo="

[app.settings]
debug = true
timeout = 30
theme = "dark"

[orm]
writer = "host=127.0.0.1 port=5432 user=pgsql password=123456 dbname=database sslmode=disable TimeZone=UTC application_name=db"
readers = [
"host=127.0.0.1 port=5432 user=pgsql password=123456 dbname=database sslmode=disable TimeZone=UTC application_name=db",
]
driver = "postgres"
disable_auto_report = false
enable_debug = false
auto_create_database = true
init_sqls = ["CREATE EXTENSION citext"]

[http]
address = "0.0.0.0"
port = 8000
base_url = "http://127.0.0.1:8000"

[logger]
output = "stdout"
level = "info"
format = "json"
disable_report_caller = false

[cache]
enable = false
type = "redis"
dsn = "redis://127.0.0.1:6379/0"
```

Below is an example of the configuration file `dev.toml`, demonstrating how to set up application configuration, database, cache, HTTP, and logging modules.

```toml
[app]
name = "App Name"
tags = ["golang", "viper", "config"]
abc = "ffefef"
ccc = 12345
secret_key = "c2VjcmV0X2tleQo="

[app.settings]
debug = true
timeout = 30
theme = "dark"

[orm]
writer = "host=127.0.0.1 port=5432 user=pgsql password=123456 dbname=database sslmode=disable TimeZone=UTC application_name=db"
readers = [
"host=127.0.0.1 port=5432 user=pgsql password=123456 dbname=database sslmode=disable TimeZone=UTC application_name=db",
]
driver = "postgres"
disable_auto_report = false
enable_debug = false
auto_create_database = true
init_sqls = ["CREATE EXTENSION citext"]

[http]
address = "0.0.0.0"
port = 8000
base_url = "http://127.0.0.1:8000"

[logger]
output = "stdout"
level = "info"
format = "json"
disable_report_caller = false

[cache]
enable = false
type = "redis"
dsn = "redis://127.0.0.1:6379/0"
```

---
