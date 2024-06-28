package e2gin

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/e2u/e2util/e2exec"
	"github.com/e2u/e2util/e2hash/e2md5"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var etagCache = sync.Map{}

func fileExists(webFs fs.FS, path string) bool {
	return e2exec.OnlyError(webFs.Open(path)) == nil
}

func AddStaticFs(staticFs fs.FS, r *gin.Engine, httpPath string) {
	httpPath = filepath.Clean(httpPath)

	requestPathFunc := func(c *gin.Context) string {
		requestPath := strings.TrimLeft(c.Request.RequestURI, httpPath)
		if before, _, ok := strings.Cut(requestPath, "?"); ok {
			requestPath = before
			return requestPath
		}
		return requestPath
	}

	r.Use(func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.RequestURI, httpPath) {
			c.Next()
			return
		}

		requestFile := requestPathFunc(c)
		if strings.HasSuffix(c.Request.RequestURI, "/") && requestFile == "" && fileExists(staticFs, "index.html") {
			c.Request.RequestURI += "index.html"
			c.Next()
			return
		}

		if requestFile == "" || strings.HasSuffix(requestFile, "/") {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		var etag string
		v, ok := etagCache.Load(requestFile)
		if ok && v != nil {
			etag = v.(string)
		} else {
			b, err := readFileContent(staticFs, requestFile)
			if err != nil {
				logrus.Errorf("read file error=%v", err)
				c.Next()
				return
			}
			etag = e2md5.MD5HexString(b)
			etagCache.Store(requestFile, etag)
		}

		if MatchEtag(c, []byte(etag)) {
			return
		}
	}).StaticFS(httpPath, http.FS(staticFs))
}

func readFileContent(staticFs fs.FS, fileName string) ([]byte, error) {
	if fileName == "" {
		return nil, fs.ErrNotExist
	}
	f, err := staticFs.Open(fileName)
	if err != nil {
		logrus.Errorf("open file error=%v, filename=%v", err, fileName)
		return nil, err
	}
	defer e2exec.MustClose(f)
	b, err := io.ReadAll(f)
	if err != nil {
		logrus.Errorf("read file error=%v, filename=%v", err, fileName)
		return nil, err
	}
	return b, nil
}

func AddEmbedStaticFs(efs embed.FS, r *gin.Engine, httpPath string) {
	staticFs, _ := fs.Sub(efs, ".")
	AddStaticFs(staticFs, r, httpPath)
}

func MatchEtag(c *gin.Context, data []byte) bool {
	etag := e2md5.HeadTailHex(data)
	c.Header("Cache-Control", "public, max-age=31536000")
	c.Header("ETag", etag)

	if match := c.GetHeader("If-None-Match"); match != "" && match == etag {
		c.AbortWithStatus(http.StatusNotModified)
		return true
	}
	return false
}
