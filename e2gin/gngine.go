package e2gin

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Option struct {
	Root             string
	DisabledPprof    bool
	PprofPathPrefix  string
	DisableHealth    bool
	HealthPathPrefix string
	Engine           *gin.Engine
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
	router.Use(gin.Recovery())
	router.Use(gzip.Gzip(gzip.DefaultCompression))

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

	return router
}
