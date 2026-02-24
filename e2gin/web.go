package e2gin

import (
	"crypto/md5"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/e2u/e2util/e2hash"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var etagCache = sync.Map{}

// Pre-compiled regex patterns for cleanHttpPath
var (
	cleanHttpPathRe1 = regexp.MustCompile(`\\+`)
	cleanHttpPathRe2 = regexp.MustCompile(`^[./\\]+`)
	// safePathRe validates that path doesn't contain dangerous characters
	safePathRe = regexp.MustCompile(`^[a-zA-Z0-9._~!$&'()*+,;=:@/-]+$`)
)

// isSafePath validates that the path is safe (no path traversal)
func isSafePath(path string) bool {
	// Reject paths containing ..
	if strings.Contains(path, "..") {
		return false
	}
	// Must match safe path pattern
	return safePathRe.MatchString(path)
}

// getContentType detects content type from file extension
func getContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".eot":
		return "application/vnd.ms-fontobject"
	case ".otf":
		return "font/otf"
	default:
		return "application/octet-stream"
	}
}

func cleanHttpPath(s string) string {
	httpPath := filepath.Clean(s)
	httpPath = cleanHttpPathRe1.ReplaceAllString(httpPath, "/")
	httpPath = cleanHttpPathRe2.ReplaceAllString(httpPath, "/")

	if !strings.HasPrefix(httpPath, "/") {
		httpPath = "/" + httpPath
	}
	// Ensure path does not end with / (except for root /)
	httpPath = strings.TrimSuffix(httpPath, "/")
	if httpPath == "" {
		httpPath = "/"
	}
	return httpPath
}

// registerStaticFiles registers static file handlers with support for hot reload in dev mode
// In development mode, files are served from localPath directly to support hot reload
// In release mode, files are served from the embedded fs.FS
func registerStaticFiles(r *gin.Engine, staticFs fs.FS, httpPath string, localPath string) {
	rg := r.Group(httpPath, cacheMiddleware())

	// Collect all file paths first
	var files []string
	err := fs.WalkDir(staticFs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		logrus.Errorf("registerStaticFiles error walking directory: %v", err)
		return
	}

	// Determine if we're in development mode with local path
	isDevMode := gin.Mode() != gin.ReleaseMode && localPath != ""

	// Register handlers for each file
	for _, filePath := range files {
		// Capture loop variables
		fp := filePath
		routePath := "/" + strings.ReplaceAll(fp, "\\", "/")

		if isDevMode {
			// Development mode: serve from local disk for hot reload
			rg.GET(routePath, func(c *gin.Context) {
				// Security check
				if !isSafePath(fp) {
					c.AbortWithStatus(http.StatusForbidden)
					return
				}

				fullPath := filepath.Join(localPath, fp)

				// Check if file exists
				info, err := os.Stat(fullPath)
				if err != nil || info.IsDir() {
					c.AbortWithStatus(http.StatusNotFound)
					return
				}

				// Read and serve file
				content, err := os.ReadFile(fullPath)
				if err != nil {
					logrus.Errorf("Failed to read file %s: %v", fullPath, err)
					c.AbortWithStatus(http.StatusInternalServerError)
					return
				}

				contentType := getContentType(fp)
				c.Data(http.StatusOK, contentType, content)
			})
		} else {
			// Release mode: serve from embedded FS
			rg.StaticFileFS(routePath, fp, http.FS(staticFs))
		}
	}
}

func settingEtag(staticFs fs.FS, httpPath string) {
	logrus.Infof("setting Etag for %s", httpPath)
	_ = fs.WalkDir(staticFs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logrus.Errorf("settingEtag: walk error: %v", err)
			return nil // Continue walking despite error
		}
		if d == nil || d.IsDir() {
			return nil
		}
		f, fErr := staticFs.Open(path)
		if fErr != nil {
			logrus.Errorf("settingEtag: open file, error=%v", fErr)
			return fErr
		}
		defer f.Close()

		// Use streaming hash to avoid loading entire file into memory
		etagHash, err := e2hash.HashHexReader(f, md5.New)
		if err != nil {
			logrus.Errorf("settingEtag: hash file, error=%v", err)
			return err
		}

		// Normalize cache key to match request URL path
		// Ensure consistent path format: no trailing slash on base path, / separator before file
		httpPath = strings.TrimSuffix(httpPath, "/")
		cacheKey := "/" + strings.TrimPrefix(strings.ReplaceAll(filepath.Join(httpPath, path), "\\", "/"), "/")
		logrus.Debugf("cacheKey=%s, etag hash=%v", cacheKey, etagHash)
		etagCache.Store(cacheKey, etagHash)
		return nil
	})
}

func cacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Next()
			return
		}

		// Normalize request path for cache lookup
		requestPath := c.Request.URL.Path
		if requestPath == "" {
			requestPath = "/"
		}

		etag, ok := etagCache.Load(requestPath)
		if !ok {
			c.Next()
			return
		}

		etagStr := etag.(string)
		// Handle If-None-Match header (may be quoted or unquoted)
		if match := c.GetHeader("If-None-Match"); match != "" {
			// Strip quotes from both sides for comparison
			match = strings.Trim(match, `"`)
			if match == etagStr {
				c.AbortWithStatus(http.StatusNotModified)
				return
			}
		}
		c.Header("ETag", fmt.Sprintf(`"%s"`, etagStr))
		c.Next()
	}
}
