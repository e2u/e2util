# e2gin

`e2gin` 是 `e2util` 的 Web 框架增强包，基于 [Gin](https://github.com/gin-gonic/gin) 提供了一套完整的 Web 应用开发工具，支持 SPA/非 SPA 应用、静态文件服务、动态模板渲染等功能。

`e2gin` is a Gin-based web helper: SPA or multi-page static files, HTML template hot reload, health/pprof, gzip, and structured JSON responses. Subpackages: `component` (pagination), `middlewares`, `req`, `resp`.

---

## 安装 / Installation

```bash
go get github.com/e2u/e2util/e2gin
```

---

## 快速开始 / Quick start

### 基础用法 / Basic usage

```go
package main

import (
    "github.com/e2u/e2util/e2gin"
    "embed"
)

//go:embed static/*
var staticFS embed.FS

func main() {
    opt := &e2gin.Option{
        StaticFiles: []*e2gin.StaticFiles{
            {
                FS:       staticFS,
                HttpPath: "/",
            },
        },
    }

    eng := e2gin.DefaultEngine(opt)
    e2gin.StartAndStopHttp(eng, "0.0.0.0", 8080, func() {})
}
```

---

## 核心功能 / Features

### 1. 静态文件服务 / Static files

支持 SPA（单页应用）和非 SPA 应用的路由处理：

| 场景 | 请求路径 | 存在文件 | 返回内容 |
|------|----------|----------|----------|
| **SPA** | `/login` | `index.html` | `index.html`（前端路由处理） |
| **非 SPA** | `/login` | `login.html` | `login.html`（独立页面） |
| **混合** | `/about` | `about.html` | `about.html`（优先） |
| **混合** | `/dashboard` | 无，但 `index.html` 存在 | `index.html`（SPA 回退） |

```go
opt := &e2gin.Option{
    StaticFiles: []*e2gin.StaticFiles{
        // 根路径静态文件
        {
            FS:        webFS,       // 嵌入的静态文件
            HttpPath:  "/",         // HTTP 路径
            LocalPath: "./static",  // 开发模式热重载路径
        },
        // 子路径静态文件
        {
            FS:        assetsFS,
            HttpPath:  "/assets",
        },
    },
}
```

**开发模式热重载**：当 `gin.Mode() != gin.ReleaseMode` 且设置了 `LocalPath`，文件变更会自动生效，无需重启服务器。

---

### 2. SPA vs 非 SPA 配置 / SPA vs multi-page

#### SPA 应用（React/Vue/Angular） / SPA apps

```go
// 项目结构：static/index.html (SPA 入口)
opt := &e2gin.Option{
    StaticFiles: []*e2gin.StaticFiles{
        {
            FS:       staticFS,
            HttpPath: "/",
            LocalPath: "./build", // 开发模式热重载
        },
    },
}
// /login, /dashboard 等路径都会返回 index.html
```

#### 非 SPA 应用（多页 HTML） / Multi-page HTML

```go
// 项目结构：
// static/
//   ├── index.html
//   ├── login.html
//   ├── dashboard.html

opt := &e2gin.Option{
    StaticFiles: []*e2gin.StaticFiles{
        {
            FS:       staticFS,
            HttpPath: "/",
            LocalPath: "./static",
        },
    },
}
// /login → login.html
// /dashboard → dashboard.html
// / → index.html
```

#### 混合应用 / Mixed apps

```go
// 部分页面独立 HTML，部分使用 SPA
opt := &e2gin.Option{
    StaticFiles: []*e2gin.StaticFiles{
        {
            FS:       staticFS,
            HttpPath: "/",
        },
    },
}
// /login.html → login.html (独立页面)
// /app/* → index.html (SPA 路由)
```

---

### 3. 动态模板渲染 / Dynamic templates

支持模板热重载（开发模式）和 HTML 压缩：

```go
opt := &e2gin.Option{
    Template: &e2gin.Template{
        FS:       templateFS,
        LocalPath: "./templates", // 开发模式热重载
        FuncMap: template.FuncMap{
            "upper": strings.ToUpper,
        },
        Option: e2gin.TemplatesOption{
            TrimTags:   true,  // 去除模板标签空白
            MinifyHTML: true,  // HTML 压缩
        },
    },
}

eng := e2gin.DefaultEngine(opt)

// 在 handler 中使用
type PageData struct {
    Title string
}

eng.GET("/", func(c *gin.Context) {
    c.HTML(200, "index.html", PageData{Title: "首页"})
})
```

---

### 4. 配置选项 / Options

```go
type Option struct {
    Root                   string           // URL 根路径，默认 "/"
    StaticFiles            []*StaticFiles   // 静态文件配置
    DisabledPprof          bool             // 禁用 Pprof
    PprofPathPrefix        string           // Pprof 路径前缀
    DisableHealth          bool             // 禁用健康检查
    DisableRecovery        bool             // 禁用 panic 恢复
    SkipLogPaths           []string         // 跳过日志记录的路径
    HealthPathPrefix       string           // 健康检查路径前缀
    Engine                 *gin.Engine      // 自定义 gin 引擎
    NoRouteProxyBackendURL string           // 反向代理后端地址
    DisableGzip            bool             // 禁用 Gzip
    LogrusLogger           *logrus.Logger   // 自定义日志
    Template               *Template        // 模板配置
}
```

---

### 5. 完整示例 / Full example

```go
package main

import (
    "embed"
    "time"

    "github.com/e2u/e2util/e2gin"
    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"
)

//go:embed web/build/*
var webFS embed.FS

//go:embed templates/*
var templateFS embed.FS

func main() {
    // 设置日志格式
    logrus.SetFormatter(&logrus.TextFormatter{
        ForceColors:   true,
        FullTimestamp: true,
    })

    opt := &e2gin.Option{
        // 静态文件：SPA 应用
        StaticFiles: []*e2gin.StaticFiles{
            {
                FS:        webFS,
                HttpPath:  "/",
                LocalPath: "./web/build", // 开发模式热重载
            },
        },

        // 模板配置
        Template: &e2gin.Template{
            FS:        templateFS,
            LocalPath: "./templates",
            Option: e2gin.TemplatesOption{
                TrimTags:   true,
                MinifyHTML: !gin.IsDebugging(),
            },
        },

        // 健康检查配置
        HealthPathPrefix: "/__app",

        // Pprof 配置（仅开发模式）
        DisabledPprof: gin.Mode() == gin.ReleaseMode,
        PprofPathPrefix: "/__app",

        // Gzip 压缩（生产环境启用）
        DisableGzip: gin.Mode() == gin.DebugMode,
    }

    eng := e2gin.DefaultEngine(opt)

    // API 路由
    api := eng.Group("/api/v1")
    {
        api.GET("/health", func(c *gin.Context) {
            c.JSON(200, gin.H{"status": "ok"})
        })
    }

    // 启动服务
    e2gin.StartAndStopHttp(eng, "0.0.0.0", 8080, func() {
        logrus.Info("Server shutting down...")
    })
}
```

---

## 测试 / Tests

运行包内所有测试：

```bash
go test ./e2gin/...
```

运行特定测试：

```bash
# SPA 路由测试
go test ./e2gin/... -run TestSPA

# 组件测试
go test ./e2gin/component/...
```

---

## 安全特性 / Security

- **路径遍历防护**：自动拒绝包含 `..` 的请求路径
- **空字节注入防护**：拒绝包含空字节的请求
- **安全路径验证**：静态文件路径使用白名单验证
- **CSP 安全头**：内置 Content Security Policy 中间件

```go
// 启用安全头
eng.Use(middlewares.DefaultSecurityHeaders())
```

---

## 开发模式 vs 生产模式 / Dev vs release

| 功能 | 开发模式 (`debug`) | 生产模式 (`release`) |
|------|-------------------|---------------------|
| 模板热重载 | 启用 | 禁用 |
| 静态文件热重载 | 启用 | 禁用 |
| HTML 压缩 | 禁用 | 启用 |
| Gzip 压缩 | 禁用 | 启用 |
| 日志级别 | Debug | Info |

切换模式：

```bash
# 开发模式
gin.SetMode(gin.DebugMode)

# 生产模式
gin.SetMode(gin.ReleaseMode)
```

---

## 依赖 / Dependencies

- [gin-gonic/gin](https://github.com/gin-gonic/gin) - Web 框架
- [sirupsen/logrus](https://github.com/sirupsen/logrus) - 日志
- [e2u/e2util](https://github.com/e2u/e2util) - 工具包

---

## License

MIT License - see the [LICENSE](../LICENSE) file for details.
