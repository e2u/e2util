package e2gin

import (
	"bytes"
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

	"github.com/e2u/e2util/e2exec"
	h "github.com/e2u/e2util/e2html"
	"github.com/e2u/e2util/e2io"
	"github.com/e2u/e2util/e2os"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

//go:embed resources/favicon.ico
var favicon []byte

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
		startPprof(eng, opt)
	}

	if !opt.DisableRecovery {
		eng.Use(gin.CustomRecovery(customRecovery))
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
					go e2io.WatchDir(file.LocalPath, func(s string, event fsnotify.Event) {
						ffs = os.DirFS(file.LocalPath)
						settingEtag(ffs, file.HttpPath)
					})
				}
			} else {
				ffs = file.FS
			}
			registerStaticFiles(eng, opt, ffs, file.HttpPath)
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

	return eng
}

func loadIndexPage(sfs []*StaticFiles) []byte {
	for _, fileName := range []string{"index.html", "index.htm"} {
		for _, sf := range sfs {
			if sf.HttpPath != "/" {
				continue
			}
			if gin.Mode() != gin.ReleaseMode && e2os.FileExists(sf.LocalPath) {
				if b, err := os.ReadFile(filepath.Join(sf.LocalPath, fileName)); err == nil {
					return b
				}
			}
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

func startPprof(eng *gin.Engine, opt *Option) {
	if opt.PprofPathPrefix == "" {
		opt.PprofPathPrefix = cleanHttpPath("/__app")
	}
	var once sync.Once
	go func() {
		once.Do(func() {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
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

// process staticFS / 301 redirect too many times issues
func noRouteStaticIndex(sfs []*StaticFiles) gin.HandlerFunc {
	indexPageByte := loadIndexPage(sfs)
	return func(c *gin.Context) {
		reqUri, _, _ := strings.Cut(c.Request.URL.String(), "?")
		if reqUri == "/index.html" || reqUri == "/" || reqUri == "" {
			c.Data(http.StatusOK, "text/html; charset=utf-8", indexPageByte)
		}
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
			proxyURL, _ := url.Parse(opt.NoRouteProxyBackendURL)
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
			}
			return
		}
	}
}

func hostPortActive(host string) bool {
	if _, err := net.DialTimeout("tcp", host, 100*time.Millisecond); err == nil {
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

func customRecovery(c *gin.Context, err any) {
	trackId := uuid.NewString()
	logrus.Errorf("Recovered %v", "8<"+strings.Repeat("-", 50))
	logrus.Errorf("TrackId %v", trackId)
	logrus.Errorf("Recovered %v", strings.Repeat("-", 50)+">8")

	dumpReq := func() string {
		var rs []string
		rs = append(rs, "\n\n")
		b, _ := httputil.DumpRequest(c.Request, false)
		for _, s := range bytes.Split(b, []byte("\n")) {
			if bytes.HasPrefix(s, []byte("Cookie")) {
				continue
			}
			rs = append(rs, string(s))
		}
		rs = append(rs, "\n\n")
		return strings.Join(rs, "\n")
	}()

	body := h.T("html", h.A("lang", "en"),
		h.T("head", h.T("title", h.Text("ServerError"))),
		h.T("body",
			h.T("h1", "Internal Server Error"),
			h.T("ul", h.Attr{"style": "list-style: none"},
				h.T("li", fmt.Sprintf("TrackId: %s", trackId)),
				h.T("li", time.Now().UTC().Format(time.RFC1123)),
				h.T("<!--", dumpReq),
			),
		),
	).String()

	c.Header("X-Track-Id", trackId)
	c.Header("Content-Type", "text/html")
	_, _ = c.Writer.WriteString(h.Doctype("html") + body)
	c.AbortWithStatus(http.StatusInternalServerError)
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
