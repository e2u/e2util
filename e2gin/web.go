package e2gin

import (
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/e2u/e2util/e2http"
	"github.com/e2u/e2util/e2io"
	"github.com/gin-gonic/gin"
)

func fileExists(webFs fs.FS, path string) bool {
	if f, err := webFs.Open(path); err == nil {
		f.Close()
		return true
	}
	return false
}

func serveFile(c *gin.Context, webFs fs.FS, path string, contentType string) {
	file, err := webFs.Open(path)
	if err != nil {
		c.String(http.StatusNotFound, "File Not Found: %s", path)
		return
	}
	defer file.Close()
	content := e2io.MustReadAll(file)
	c.Data(http.StatusOK, contentType, content)
}

func WebContent(c *gin.Context, staticFiles embed.FS, subDir string) {
	webFs, _ := fs.Sub(staticFiles, subDir)
	path := c.Request.URL.Path
	if len(path) >= 1 && strings.HasPrefix(path, "/") {
		path = path[1:]
	}

	if path == "" || !fileExists(webFs, path) {
		serveFile(c, webFs, "index.html", "text/html")
		return
	}

	ct := e2http.GetContentType(filepath.Ext(path))
	serveFile(c, webFs, path, ct)
}
