package e2gin

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	_ "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"
	"uuid"

	"github.com/e2u/e2util/e2exec"
	h "github.com/e2u/e2util/e2html"
	"github.com/e2u/e2util/e2io"
	"github.com/e2u/e2util/e2os"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//go:embed resources/favicon.ico
var favicon []byte

const (
	keyDisableGzip = "__DISABLE_GZIP__"
)

type Option struct {
	Root                   string // http url root
	StaticFiles            []*StaticFiles
	DisabledPprof          bool
	PprofPathPrefix        string
	DisableHealth          bool
	DisableRecovery        bool
	SkipLogPaths           []string
	HealthPathPrefix       string
	Engine                 *gin.Engine
	NoRouteProxyBackendURL string
	DisableGzip            bool
	LogrusLogger           *logrus.Logger
	Template               *Template
}

type Template struct {
	fs.FS
	FuncMap   template.FuncMap // or e2gin.FuncMap = template.FuncMap{"funcName":func()string{return "hello"}}
	Option    TemplatesOption
	LocalPath string // only using on dev mode
}

type StaticFiles struct {
	fs.FS
	HttpPath  string // same to local path if leave blank
	LocalPath string // only using on dev mode
}

func DefaultEngine(opt *Option) *gin.Engine {
	if opt == nil {
		opt = &Option{}
	}

	var eng *gin.Engine

	if opt.Engine == nil {
		eng = gin.New()
	} else {
		eng = opt.Engine
	}

	if opt.Template == nil {
		opt.Template = &Template{
			Option: TemplatesOption{
				TrimTags:   false,
				MinifyHTML: false,
			},
			LocalPath: "./templates",
		}
	}

	if topt := opt.Template; topt != nil {
		if topt.FS != nil {
			eng.SetHTMLTemplate(e2exec.Must(ParseTemplates(topt.FS, topt.FuncMap, topt.Option)))
		}

		if topt.LocalPath == "" {
			topt.LocalPath = "./templates"
		}

		if gin.Mode() != gin.ReleaseMode && e2os.FileExists(topt.LocalPath) {
			eng.HTMLRender = NewDynamicHTMLRender(topt.LocalPath, topt.FuncMap, topt.Option)
		}
	}

	if opt.Root == "" {
		opt.Root = "/"
	}

	if opt.LogrusLogger == nil || reflect.ValueOf(opt.LogrusLogger).IsNil() {
		eng.Use(ginrus.Ginrus(logrus.StandardLogger(), time.RFC3339Nano, false))
	} else {
		eng.Use(ginrus.Ginrus(opt.LogrusLogger, time.RFC3339Nano, false))
	}

	if !opt.DisableHealth {
		if opt.HealthPathPrefix == "" {
			opt.HealthPathPrefix = "/__app"
		}

		hg := eng.Group(opt.Root)
		{
			hg.Use(gin.LoggerWithConfig(gin.LoggerConfig{
				SkipPaths: []string{opt.Root + "/_health", "/_health"},
			}))

			hg.GET(opt.HealthPathPrefix+"/_health", func(c *gin.Context) {
				c.String(http.StatusOK, "OK")
			})

			hg.HEAD(opt.HealthPathPrefix+"/_health", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})
		}
	}

	if !opt.DisabledPprof {
		startPprof(context.Background(), eng, opt)
	}

	eng.RemoveExtraSlash = true
	eng.HandleMethodNotAllowed = true

	if len(opt.StaticFiles) > 0 {
		var watchingStatic sync.Map
		for _, file := range opt.StaticFiles {
			if file.HttpPath == "" && file.LocalPath != "" {
				file.HttpPath = cleanHttpPath(file.LocalPath)
			}
			var ffs fs.FS
			if gin.Mode() != gin.ReleaseMode && e2os.FileExists(file.LocalPath) {
				ffs = os.DirFS(file.LocalPath)
				if _, loaded := watchingStatic.LoadOrStore(file.LocalPath, struct{}{}); !loaded {
					// Capture loop variables for goroutine
					localPath := file.LocalPath
					httpPath := file.HttpPath
					go e2io.WatchDir(localPath, func(s string, event fsnotify.Event) {
						// Create new FS instance after file change
						newFs := os.DirFS(localPath)
						settingEtag(newFs, httpPath)
					})
				}
			} else {
				ffs = file.FS
			}
			registerStaticFiles(eng, ffs, file.HttpPath, file.LocalPath)
			settingEtag(ffs, file.HttpPath)
		}
	}

	// only the last one NoRoute method will be executed
	noRouteChain := []gin.HandlerFunc{
		noRouteStaticIndex(opt.StaticFiles),
		noRouteFavicon(),
		noRouteProxy(opt),
	}

	eng.NoRoute(noRouteChain...)

	if !opt.DisableGzip {
		eng.Use(gzip.Gzip(gzip.DefaultCompression))
	}

	if !opt.DisableRecovery {
		eng.Use(gin.CustomRecovery(customRecovery))
	}

	return eng
}

// loadIndexPage 从静态文件中加载 index.html
// 支持从任何配置的静态文件目录加载，不仅限于根路径
func loadIndexPage(sfs []*StaticFiles) []byte {
	return loadHTMLPage(sfs, "index")
}

// loadHTMLPage 从静态文件中加载指定名称的 HTML 文件
// 例如 name="login" 会查找 login.html 或 login.htm
func loadHTMLPage(sfs []*StaticFiles, name string) []byte {
	for _, ext := range []string{".html", ".htm"} {
		fileName := name + ext
		for _, sf := range sfs {
			// 优先从本地路径加载（开发模式）
			if gin.Mode() != gin.ReleaseMode && e2os.FileExists(sf.LocalPath) {
				if b, err := os.ReadFile(filepath.Clean(filepath.Join(sf.LocalPath, fileName))); err == nil {
					return b
				}
			}
			// 从嵌入的 FS 加载
			if f, err := sf.Open(fileName); err == nil {
				if b, rErr := io.ReadAll(f); rErr == nil {
					_ = f.Close()
					return b
				}
				_ = f.Close()
			}
		}
	}
	return nil
}

// hasHTMLPage 检查是否存在指定名称的 HTML 文件
func hasHTMLPage(sfs []*StaticFiles, name string) bool {
	for _, ext := range []string{".html", ".htm"} {
		fileName := name + ext
		for _, sf := range sfs {
			// 检查本地路径（开发模式）
			if gin.Mode() != gin.ReleaseMode && e2os.FileExists(sf.LocalPath) {
				if _, err := os.Stat(filepath.Join(sf.LocalPath, fileName)); err == nil {
					return true
				}
			}
			// 检查嵌入的 FS
			if _, err := sf.Open(fileName); err == nil {
				return true
			}
		}
	}
	return false
}

func startPprof(ctx context.Context, eng *gin.Engine, opt *Option) {
	if opt.PprofPathPrefix == "" {
		opt.PprofPathPrefix = cleanHttpPath("/__app")
	}
	var once sync.Once

	go func() {
		once.Do(func() {
			lc := net.ListenConfig{}
			listener, err := lc.Listen(ctx, "tcp", "127.0.0.1:0")
			if err != nil {
				logrus.Errorf("make tcp listen error: %v", err)
				return
			}

			port := listener.Addr().(*net.TCPAddr).Port
			logrus.Infof("pprof port: %v", port)
			pprofUrl := fmt.Sprintf("http://127.0.0.1:%d/debug/pprof", port)
			logrus.Info(pprofUrl)

			eng.GET(opt.PprofPathPrefix+"/pprof-info", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"pprof_url": pprofUrl,
					"command": []string{
						fmt.Sprintf("ssh -N -L %d:127.0.0.1:%d <ssh-host>", port, port),
						fmt.Sprintf("go tool pprof -http=:18081 http://127.0.0.1:%d/debug/pprof/profile -seconds 30", port),
					},
				})
			})

			if err := http.Serve(listener, nil); err != nil { // #nosec G114
				logrus.Infof("run pprof error: %v", err)
				return
			}
		})
	}()
}

// noRouteStaticIndex 处理静态 HTML 页面路由
// 支持 SPA 和非 SPA 应用：
//   - 非 SPA: /login → 查找并返回 login.html
//   - SPA: 如果找不到对应页面，返回 index.html 由前端路由处理
func noRouteStaticIndex(sfs []*StaticFiles) gin.HandlerFunc {
	// 预加载 index.html（用于 SPA 回退）
	indexPageByte := loadIndexPage(sfs)

	return func(c *gin.Context) {
		// 只处理 GET 和 HEAD 请求
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Next()
			return
		}

		reqUri := c.Request.URL.Path

		// 安全检查：拒绝包含路径遍历或空字节的请求
		if strings.Contains(reqUri, "..") || strings.ContainsRune(reqUri, 0) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// 排除已知的 API 路径模式
		apiPrefixes := []string{"/api/", "/__app/", "/static/", "/assets/", "/favicon.ico"}
		for _, prefix := range apiPrefixes {
			if strings.HasPrefix(reqUri, prefix) {
				c.Next()
				return
			}
		}

		// 排除文件扩展名（如 .js, .css, .png 等），但保留 .html
		if ext := filepath.Ext(reqUri); ext != "" && ext != ".html" {
			c.Next()
			return
		}

		// 检查 Accept 头，确保是浏览器请求
		accept := c.GetHeader("Accept")
		isBrowserRequest := accept == "" ||
			strings.Contains(accept, "text/html") ||
			strings.Contains(accept, "*/*")
		if !isBrowserRequest {
			c.Next()
			return
		}

		// 提取页面名称（去掉前导 / 和 .html 后缀）
		pageName := strings.TrimPrefix(reqUri, "/")
		pageName = strings.TrimSuffix(pageName, ".html")

		// 情况 1: 根路径，直接返回 index.html
		if pageName == "" || pageName == "index" {
			if len(indexPageByte) > 0 {
				c.Data(http.StatusOK, "text/html; charset=utf-8", indexPageByte)
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// 情况 2: 非根路径，首先尝试查找对应的 .html 文件（非 SPA 支持）
		// 例如 /login → 查找 login.html
		if hasHTMLPage(sfs, pageName) {
			pageContent := loadHTMLPage(sfs, pageName)
			if len(pageContent) > 0 {
				c.Data(http.StatusOK, "text/html; charset=utf-8", pageContent)
				c.Abort()
				return
			}
		}

		// 情况 3: 没有找到对应页面，如果是 SPA 则返回 index.html
		// 这样前端路由可以处理 /login 这样的路径
		if len(indexPageByte) > 0 {
			c.Data(http.StatusOK, "text/html; charset=utf-8", indexPageByte)
			c.Abort()
			return
		}

		c.Next()
	}
}

// the noRouteFavicon consider to run at last one
func noRouteFavicon() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.RequestURI == "/favicon.ico" {
			c.Header("Cache-Control", "public, max-age=3600, must-revalidate")
			c.Data(http.StatusOK, "image/x-icon", favicon)
			return
		}
	}
}

func noRouteProxy(opt *Option) gin.HandlerFunc {
	return func(c *gin.Context) {
		if opt.NoRouteProxyBackendURL != "" {
			proxyURL, err := url.Parse(opt.NoRouteProxyBackendURL)
			if err != nil {
				logrus.Errorf("invalid proxy URL: %v", err)
				c.AbortWithStatus(http.StatusBadGateway)
				return
			}
			if hostPortActive(proxyURL.Host) {
				proxy := httputil.NewSingleHostReverseProxy(proxyURL)
				proxy.FlushInterval = time.Millisecond * 100
				proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
					slog.Error("Error proxying request", "error", err)
				}
				proxy.ModifyResponse = func(resp *http.Response) error {
					if resp.StatusCode == http.StatusOK {
						resp.Header.Add("X-Content-Source", "proxy")
					}
					return nil
				}
				proxy.ServeHTTP(c.Writer, c.Request)
				return
			}
			// Proxy configured but backend not available
			c.AbortWithStatus(http.StatusBadGateway)
			return
		}
		// No proxy configured, return 404
		c.AbortWithStatus(http.StatusNotFound)
	}
}

func hostPortActive(host string) bool {
	var d net.Dialer
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	if _, err := d.DialContext(ctx, "tcp", host); err == nil {
		return true
	}
	return false
}

func StartAndStopHttp(eng *gin.Engine, address string, port int, stop func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		logrus.Infof("Server started. Press Ctrl+C to stop.")
		if err := eng.Run(fmt.Sprintf("%s:%d", address, port)); err != nil {
			logrus.Fatal(err)
		}
	}()
	<-sigChan
	fmt.Println("Received SIGINT or SIGTERM. Shutting down...")
	stop()
	os.Exit(0)
}

func customRecovery(c *gin.Context, msg any) {
	c.Set(keyDisableGzip, true)

	trackId := uuid.New().String()

	dumpReq := func() string {
		var rs []string
		rs = append(rs, "\n")
		b, _ := httputil.DumpRequest(c.Request, false)
		for s := range bytes.SplitSeq(b, []byte("\n")) {
			if bytes.HasPrefix(s, []byte("Cookie")) {
				continue
			}
			rs = append(rs, string(s))
		}
		rs = append(rs, "\n")
		return strings.Join(rs, "\n")
	}()

	logrus.Errorf("Recovered %v", "8<"+strings.Repeat("-", 50))
	logrus.Errorf("TrackId %v", trackId)
	logrus.Errorf("Error: %v", msg)
	logrus.Errorf("DumpReq: %v", dumpReq)
	logrus.Errorf("Recovered %v", strings.Repeat("-", 50)+">8")

	msgStr := func() string {
		switch v := msg.(type) {
		case error:
			return v.Error()
		default:
			return fmt.Sprintf("%v", msg)
		}
	}

	body := h.T("body",
		h.T("h1", "Internal Server Error"),
		h.T("ul", h.Attr{"style": "list-style: none"},
			h.T("li", fmt.Sprintf("TrackId: %s", trackId)),
			h.T("li", time.Now().UTC().Format(time.RFC1123)),
		),
		h.T("pre", fmt.Sprintf("Error Message: %s", msgStr())),
		// h.T("pre", dumpReq),
	)

	html := h.T("html", h.A("lang", "en"),
		h.T("head", h.T("title", h.Text("ServerError"))),
		body,
	)

	c.Header("X-Track-SessionId", trackId)
	c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(h.Doctype("html")+html.String()))
	c.Abort()
}

func errorPage(title string, err error) string {
	return h.T("html", h.A("lang", "en"),
		h.T("head", h.T("title", h.Text("Error"))),
		h.T("body",
			h.T("h1", title),
			h.T("pre", h.Text(err.Error())),
		),
	).String()
}
