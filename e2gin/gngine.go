package e2gin

import (
	"bytes"
	"embed"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	_ "net/http/pprof"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	h "github.com/e2u/e2util/e2html"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Option struct {
	Root                   string
	StaticFiles            *StaticFiles
	DisabledPprof          bool
	PprofPathPrefix        string
	DisableHealth          bool
	SkipLogPaths           []string
	HealthPathPrefix       string
	Engine                 *gin.Engine
	NoRouteProxyBackendURL string
	DisableGzip            bool
}

type StaticFiles struct {
	embed.FS
	SubDir string // example: "web/build"
}

func DefaultEngine(opt *Option) *gin.Engine {
	if opt == nil {
		opt = &Option{}
	}

	var router *gin.Engine

	if opt.Engine == nil {
		router = gin.New()
	} else {
		router = opt.Engine
	}

	hg := router.Group(opt.Root)
	{
		hg.Use(gin.LoggerWithConfig(gin.LoggerConfig{
			SkipPaths: []string{opt.Root + "/_health", "/_health"},
		}))

		if !opt.DisableHealth {
			hg.GET(opt.HealthPathPrefix+"/_health", func(c *gin.Context) {
				c.String(http.StatusOK, "OK")
			})

			hg.HEAD(opt.HealthPathPrefix+"/_health", func(c *gin.Context) {
				c.Status(http.StatusOK)
			})
		}
	}

	router.Use(ginrus.Ginrus(logrus.StandardLogger(), time.RFC3339, false))
	if !opt.DisableGzip {
		router.Use(gzip.Gzip(gzip.DefaultCompression))
	}
	router.Use(gin.CustomRecovery(customRecovery))
	router.RemoveExtraSlash = true
	router.HandleMethodNotAllowed = true

	if opt.DisabledPprof {
		return router
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

			router.GET(opt.PprofPathPrefix+"/pprof-info", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"pprof_url": pprofUrl,
					"command": []string{
						fmt.Sprintf("ssh -N -L %d:127.0.0.1:%d <ssh-host>", port, port),
						fmt.Sprintf("go tool pprof -http=:18081 http://127.0.0.1:%d/%s/debug/pprof/profile -seconds 30", port, opt.PprofPathPrefix),
					},
				})
			})

			if err := http.Serve(listener, nil); err != nil { // #nosec G114
				logrus.Infof("run pprof error: %v", err)
				return
			}
		})
	}()

	router.NoRoute(func(c *gin.Context) {
		var send bool
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
						send = true
					}
					return nil
				}
				proxy.ServeHTTP(c.Writer, c.Request)
			}
		}

		if opt.StaticFiles != nil && !send {
			c.Header("X-Content-Source", "fs")
			WebContent(c, opt.StaticFiles.FS, opt.StaticFiles.SubDir)
		}
	})

	return router
}

func hostPortActive(host string) bool {
	if _, err := net.DialTimeout("tcp", host, 100*time.Millisecond); err == nil {
		return true
	}
	return false
}

func StartAndStop(start func(), stop func()) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		fmt.Println("Server started. Press Ctrl+C to stop.")
		start()
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

	body := h.T("html", h.A{"lang": "en"},
		h.T("head", h.T("title", h.Text("ServerError"))),
		h.T("body",
			h.T("h1", "Internal Server Error"),
			h.T("ul", h.A{"style": "list-style: none"},
				h.T("li", fmt.Sprintf("TrackId: %s", trackId)),
				h.T("li", time.Now().UTC().Format(time.RFC1123)),
				h.T("<!--", dumpReq),
			),
		),
	).String()

	c.Writer.WriteHeader(http.StatusInternalServerError)
	c.Writer.Header().Set("X-Track-Id", trackId)
	c.Writer.Header().Set("Content-Type", "text/html")
	_, _ = c.Writer.WriteString(h.Doctype("html") + body)
}
