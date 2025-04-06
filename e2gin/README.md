---

# e2gin Documentation

## 項目概覽 / Project Overview

`e2gin` 是 `e2util` 下的子包，位於 `github.com/e2u/e2util/e2gin`，提供了一組增強 `gin` Web 框架的工具函數和功能，主要包含以下模塊：
1. **動態模板渲染**：支持模板文件的動態加載與監控，適用於開發和生產環境。
2. **靜態文件服務**：提供靜態文件服務，支持 ETag 緩存和壓縮。
3. **應用配置與路由**：簡化 `gin` 引擎配置，支持健康檢查、Pprof 調試和反向代理等功能。
此包依賴 `github.com/gin-gonic/gin`、`github.com/sirupsen/logrus` 等庫，並整合了 `e2util` 的多個工具，提供高效的 Web 應用開發支持。

`e2gin` is a sub-package under `e2util`, located at `github.com/e2u/e2util/e2gin`, offering a set of tools and enhancements for the `gin` web framework. It includes the following main modules:
1. **Dynamic Template Rendering**: Supports dynamic loading and monitoring of template files, suitable for both development and production environments.
2. **Static File Serving**: Provides static file serving with ETag caching and compression support.
3. **Application Configuration and Routing**: Simplifies `gin` engine configuration, supporting health checks, Pprof debugging, and reverse proxy features.
This package depends on `github.com/gin-gonic/gin`, `github.com/sirupsen/logrus`, and integrates multiple `e2util` tools, offering efficient support for web application development.

---

## 使用方法 / Usage

### 1. 配置並啟動動態模板渲染 / Configure and Start Dynamic Template Rendering

```go
package main

import (
"github.com/e2u/e2util/e2gin"
"github.com/gin-gonic/gin"
)

func main() {
// 初始化 gin 引擎 / Initialize gin engine
r := gin.Default()

// 配置動態模板渲染 / Configure dynamic template rendering
render := e2gin.NewDynamicHTMLRender("./templates")
r.HTMLRender = render

// 定義路由 / Define route
r.GET("/", func(c *gin.Context) {
c.HTML(200, "index.html", nil)
})
}
```

### 2. 配置並運行默認引擎 / Configure and Run Default Engine

```go
package main

import (
"github.com/e2u/e2util/e2gin"
)

func main() {
// 配置默認引擎 / Configure default engine
opt := &e2gin.Option{
StaticFiles: []*e2gin.StaticFiles{{LocalPath: "./static", HttpPath: "/static"}},
}
eng := e2gin.DefaultEngine(opt)

// 啟動服務器 / Start server
e2gin.StartAndStopHttp(eng, "0.0.0.0", 8080, func() {})
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
go get -u github.com/e2u/e2util/e2gin
```

3. **驗證安裝**
在代碼中導入 `github.com/e2u/e2util/e2gin`，運行 `go build` 或 `go run`，若無錯誤則安裝成功。
Import `github.com/e2u/e2util/e2gin` in your code and run `go build` or `go run`. Success indicates proper installation.

---

## 功能描述 / Features

- **動態模板渲染**：`NewDynamicHTMLRender` 提供模板文件監控與動態重載，支持自定義函數映射和 HTML 壓縮。
**Dynamic Template Rendering**: `NewDynamicHTMLRender` offers template file monitoring and dynamic reloading, supporting custom function maps and HTML minification.

- **靜態文件服務**：`registerStaticFiles` 和 `settingEtag` 提供靜態文件服務，支持 ETag 緩存和 Gzip 壓縮。
**Static File Serving**: `registerStaticFiles` and `settingEtag` provide static file serving with ETag caching and Gzip compression.

- **引擎配置**：`DefaultEngine` 配置 `gin` 引擎，支持健康檢查、Pprof、反向代理和靜態文件路由。
**Engine Configuration**: `DefaultEngine` configures the `gin` engine, supporting health checks, Pprof, reverse proxy, and static file routing.

- **異常恢復**：`customRecovery` 提供自定義異常恢復，生成帶追蹤 ID 的錯誤頁面。
**Panic Recovery**: `customRecovery` provides custom panic recovery, generating error pages with tracking IDs.

- **模板解析**：`ParseTemplates` 解析模板文件，支持標籤修剪和 HTML 壓縮，提供靈活的配置選項。
**Template Parsing**: `ParseTemplates` parses template files, supporting tag trimming and HTML minification with flexible configuration options.

- **服務啟停**：`StartAndStopHttp` 提供服務啟動與信號處理，支持優雅關閉。
**Service Start/Stop**: `StartAndStopHttp` provides server startup and signal handling, supporting graceful shutdown.

---
