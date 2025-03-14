package e2gin

import (
	"crypto/md5"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/e2u/e2util/e2hash"
	"github.com/gin-gonic/contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var etagCache = sync.Map{}

func cleanHttpPath(s string) string {
	httpPath := filepath.Clean(s)
	re1 := regexp.MustCompile(`\\+`)
	httpPath = re1.ReplaceAllString(httpPath, "/")

	re2 := regexp.MustCompile(`^[./\\]+`)
	httpPath = re2.ReplaceAllString(httpPath, "/")

	if !strings.HasPrefix(httpPath, "/") {
		httpPath = "/" + httpPath
	}
	return httpPath
}

func registerStaticFiles(r *gin.Engine, opt *Option, staticFs fs.FS, httpPath string) {
	rg := r.Group(httpPath, cacheMiddleware())
	if !opt.DisableGzip {
		rg.Use(gzip.Gzip(gzip.DefaultCompression))
	}
	httpFS := http.FS(staticFs)
	err := fs.WalkDir(staticFs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		rg.StaticFileFS(path, path, httpFS)
		return err
	})
	if err != nil {
		logrus.Errorf("registerStaticFiles error=%v", err)
	}
}

func settingEtag(staticFs fs.FS, httpPath string) {
	logrus.Infof("setting Etag for %s", httpPath)
	_ = fs.WalkDir(staticFs, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		f, fErr := staticFs.Open(path)
		if fErr != nil {
			logrus.Errorf("settingEtag: read file, error=%v", err)
			return fErr
		}
		b, _ := io.ReadAll(f)
		cacheKey := strings.ReplaceAll(filepath.Join(httpPath, path), "\\", "/")
		etagHash := e2hash.HashHex(b, md5.New)
		logrus.Debugf("cacheKey=%s, etag hash=%v", cacheKey, etagHash)
		etagCache.Store(cacheKey, etagHash)
		_ = f.Close()
		return nil
	})
}

func cacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Next()
			return
		}
		etag, ok := etagCache.Load(c.Request.URL.Path)
		if !ok {
			c.Next()
			return
		}
		if match := c.GetHeader("If-None-Match"); match != "" && match == etag.(string) {
			c.AbortWithStatus(http.StatusNotModified)
			return
		}
		c.Header("ETag", `"`+etag.(string)+`"`)
		c.Next()
	}
}
