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
	router.Use(gzip.Gzip(gzip.DefaultCompression))
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
						fmt.Sprintf("go tool pprof -http=:18081 http://127.0.0.1:%d/%s/debug/pprof/profile -seconds 30", opt.PprofPathPrefix, port),
					},
				})
			})

			if err := http.Serve(listener, nil); err != nil {
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

	outputHtml := []string{
		`<!DOCTYPE html>`,
		`<html lang="en">`,
		`<head><title>ServerError</title></head>`,
		`<body>`,
		`<h1>Internal Server Error</h1>`,
		`<ul style="list-style: none">`,
		fmt.Sprintf(`<li>TrackId: %s</li>`, trackId),
		fmt.Sprintf(`<li>%s</li>`, time.Now().UTC().Format(time.RFC1123)),
		`</ul>`,
		`<!--`,
		func() string {
			var rs []string
			rs = append(rs, "\n\n")
			b, _ := httputil.DumpRequest(c.Request, false)
			for _, s := range bytes.Split(b, []byte("\n")) {
				if bytes.HasPrefix(s, []byte("Cookie")) {
					continue
				}
				rs = append(rs, fmt.Sprintf("%s", string(s)))
			}
			rs = append(rs, "\n\n")
			return strings.Join(rs, "\n")
		}(),
		`-->`,
		`</body>`,
		`</html>`,
	}

	c.Writer.WriteHeader(http.StatusInternalServerError)
	c.Writer.Header().Set("X-Track-Id", trackId)
	c.Writer.Header().Set("Content-Type", "text/html")
	c.Writer.WriteString(strings.Join(outputHtml, ""))

}
