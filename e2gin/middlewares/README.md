---

# middlewares Documentation

## 項目概覽 / Project Overview

`middlewares` 是 `e2gin` 子包的一部分，位於 `github.com/e2u/e2util/e2gin/middlewares`，提供基於 `gin` 的中間件功能，主要包含兩大模塊：
1. **請求日誌記錄**：記錄 HTTP 請求和響應的詳細信息，適用於調試和監控。
2. **安全頭設置**：配置內容安全策略（CSP）和 HTTP 安全頭，增強 Web 應用安全性。
此包依賴 `github.com/gin-gonic/gin` 和 `github.com/sirupsen/logrus`，提供靈活且可擴展的解決方案。

`middlewares` is a sub-package of `e2gin`, located at `github.com/e2u/e2util/e2gin/middlewares`, offering `gin`-based middleware functions with two main modules:
1. **Request Logging**: Logs detailed information about HTTP requests and responses, suitable for debugging and monitoring.
2. **Security Headers**: Configures Content Security Policy (CSP) and HTTP security headers to enhance web application security.
This package depends on `github.com/gin-gonic/gin` and `github.com/sirupsen/logrus`, providing flexible and extensible solutions.

---

## 使用方法 / Usage

### 1. 記錄請求和響應 / Log Request and Response

```go
package main

import (
"github.com/e2u/e2util/e2gin/middlewares"
"github.com/gin-gonic/gin"
"github.com/sirupsen/logrus"
)

func main() {
// 初始化 gin 和 logrus / Initialize gin and logrus
r := gin.Default()
logger := logrus.New()

// 使用請求日誌中間件 / Use request logging middleware
r.Use(middlewares.RequestLoggingMiddleware(logger))
r.GET("/", func(c *gin.Context) {
c.JSON(200, gin.H{"message": "hello"})
})
}
```

### 2. 配置安全頭 / Configure Security Headers

```go
package main

import (
"github.com/e2u/e2util/e2gin/middlewares"
"github.com/gin-gonic/gin"
)

func main() {
// 初始化 gin / Initialize gin
r := gin.Default()

// 使用默認安全頭中間件 / Use default security headers middleware
r.Use(middlewares.DefaultSecurityHeaders())
r.GET("/", func(c *gin.Context) {
c.String(200, "secure page")
})
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
go get -u github.com/e2u/e2util/e2gin/middlewares
```

3. **驗證安裝**
在代碼中導入 `github.com/e2u/e2util/e2gin/middlewares`，運行 `go build` 或 `go run`，若無錯誤則安裝成功。
Import `github.com/e2u/e2util/e2gin/middlewares` in your code and run `go build` or `go run`. Success indicates proper installation.

---

## 功能描述 / Features

- **請求日誌記錄**：`RequestLoggingMiddleware` 記錄請求方法、路徑、查詢參數、請求體和響應體，通過 `logrus` 輸出。
**Request Logging**: `RequestLoggingMiddleware` logs request method, path, query parameters, request body, and response body via `logrus`.

- **安全頭設置**：`SecurityHeaders` 配置 CSP 和其他安全頭（如 `X-Frame-Options`），支持自定義資源來源。
**Security Headers**: `SecurityHeaders` configures CSP and other security headers (e.g., `X-Frame-Options`), supporting custom resource sources.

- **默認安全配置**：`DefaultSecurityHeaders` 提供預設的安全頭配置，適用於常見場景。
**Default Security Config**: `DefaultSecurityHeaders` provides a preset security header configuration for common scenarios.

- **錯誤處理**：針對無效請求體或 JSON 解析失敗返回錯誤，確保中間件健壯性。
**Error Handling**: Returns errors for invalid request bodies or JSON parsing failures, ensuring middleware robustness.

---
