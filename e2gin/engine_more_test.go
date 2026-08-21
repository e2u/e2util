package e2gin

import (
	"bytes"
	"errors"
	"html/template"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUntil(t *testing.T) {
	assert.Equal(t, []int{1, 2, 3}, until(1, 4))
	assert.Nil(t, until(3, 3))
}

func TestErrorPage(t *testing.T) {
	got := errorPage("parse template error", errors.New("boom"))
	assert.Contains(t, got, "parse template error")
	assert.Contains(t, got, "boom")
	assert.Contains(t, got, "<html")
}

func TestStaticHasFile(t *testing.T) {
	assert.False(t, staticHasFile(nil, "x"))
	fsys := fstest.MapFS{"favicon.ico": {Data: []byte("ico")}}
	assert.True(t, staticHasFile(fsys, "favicon.ico"))
	assert.False(t, staticHasFile(fsys, "missing.ico"))
}

func TestCleanHttpPath(t *testing.T) {
	assert.Equal(t, "/foo/bar", cleanHttpPath(`\\foo\\bar`))
	assert.Equal(t, "/foo", cleanHttpPath("foo/"))
	assert.Equal(t, "/", cleanHttpPath(""))
	assert.Equal(t, "/a", cleanHttpPath("./a"))
}

func TestMinifyHTMLPreservesPre(t *testing.T) {
	in := "<div>\n  hello\n</div><pre>\n  keep\n</pre><textarea>\n  also\n</textarea>"
	got := minifyHTML(in)
	assert.Contains(t, got, "<div>hello</div>")
	assert.Contains(t, got, "<pre>\n  keep\n</pre>")
	assert.Contains(t, got, "<textarea>\n  also\n</textarea>")
}

func TestParseTemplatesFuncMapAndBadTemplate(t *testing.T) {
	fsys := fstest.MapFS{
		"ok.html":  {Data: []byte(`{{add 1 2}}`)},
		"bad.html": {Data: []byte(`{{`)},
	}
	tmpl, err := ParseTemplates(fsys, template.FuncMap{"greet": func() string { return "hi" }})
	require.NoError(t, err)
	require.NotNil(t, tmpl)

	var buf bytes.Buffer
	require.NoError(t, tmpl.ExecuteTemplate(&buf, "ok.html", nil))
	assert.Equal(t, "3", buf.String())

	buf.Reset()
	require.NoError(t, tmpl.ExecuteTemplate(&buf, "bad.html", nil))
	assert.Contains(t, buf.String(), "parse template error")
}

func TestDefaultEngineNilOptionAndHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eng := DefaultEngine(nil)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/__app/_health", nil)
	eng.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestSkipLogPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var buf bytes.Buffer
	logger := logrus.New()
	logger.SetOutput(&buf)
	eng := DefaultEngine(&Option{
		DisabledPprof:   true,
		DisableHealth:   true,
		DisableGzip:     true,
		LogrusLogger:    logger,
		SkipLogPaths:    []string{"/skip"},
		DisableRecovery: true,
	})
	eng.GET("/skip", func(c *gin.Context) { c.String(http.StatusOK, "s") })
	eng.GET("/keep", func(c *gin.Context) { c.String(http.StatusOK, "k") })

	for _, path := range []string{"/skip", "/keep"} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, path, nil)
		eng.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
	logs := buf.String()
	assert.NotContains(t, logs, `path=/skip`)
	assert.Contains(t, logs, `/keep`)
}

func TestGzipOnHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eng := DefaultEngine(&Option{DisabledPprof: true})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/__app/_health", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	eng.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRecoveryPanicErrorValue(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eng := DefaultEngine(&Option{DisabledPprof: true, DisableHealth: true, DisableGzip: true})
	eng.GET("/boom", func(c *gin.Context) { panic(errors.New("typed")) })
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/boom", nil)
	eng.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "typed")
}

func TestNoRouteProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("no backend is 404", func(t *testing.T) {
		eng := gin.New()
		eng.NoRoute(noRouteProxy(&Option{}))
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/missing", nil)
		eng.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("invalid proxy url is 502", func(t *testing.T) {
		eng := gin.New()
		eng.NoRoute(noRouteProxy(&Option{NoRouteProxyBackendURL: "http://["}))
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/x", nil)
		eng.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})

	t.Run("dead backend is 502", func(t *testing.T) {
		eng := gin.New()
		eng.NoRoute(noRouteProxy(&Option{NoRouteProxyBackendURL: "http://127.0.0.1:1"}))
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/x", nil)
		eng.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadGateway, w.Code)
	})
}

func TestHostPortActive(t *testing.T) {
	assert.False(t, hostPortActive("127.0.0.1:1"))
}

func TestRegisterStaticFilesAndETag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eng := gin.New()
	fsys := fstest.MapFS{
		"hello.txt": {Data: []byte("hello")},
	}
	registerStaticFiles(eng, fsys, "/static", "")
	settingEtag(fsys, "/static")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/static/hello.txt", nil)
	eng.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "hello", w.Body.String())
	etag := w.Header().Get("ETag")
	require.NotEmpty(t, etag)

	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/static/hello.txt", nil)
	req2.Header.Set("If-None-Match", strings.Trim(etag, `"`))
	eng.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusNotModified, w2.Code)
}

func TestRegisterStaticFilesDevLocalPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "local.txt"), []byte("from-disk"), 0o600))
	fsys := os.DirFS(dir)
	eng := gin.New()
	registerStaticFiles(eng, fsys, "/files", dir)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/files/local.txt", nil)
	eng.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "from-disk", w.Body.String())
}

func TestLoadAndHasHTMLPageFromLocalPath(t *testing.T) {
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "about.html"), []byte("<html>about</html>"), 0o600))
	sfs := []*StaticFiles{{FS: os.DirFS(dir), HttpPath: "/", LocalPath: dir}}
	assert.True(t, hasHTMLPage(sfs, "about"))
	assert.False(t, hasHTMLPage(sfs, "missing"))
	assert.Equal(t, []byte("<html>about</html>"), loadHTMLPage(sfs, "about"))
}

func TestDefaultEngineUsesCustomEngine(t *testing.T) {
	gin.SetMode(gin.TestMode)
	custom := gin.New()
	eng := DefaultEngine(&Option{Engine: custom, DisabledPprof: true, DisableHealth: true, DisableGzip: true})
	assert.Equal(t, custom, eng)
}

func TestNoRouteFaviconViaChain(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eng := gin.New()
	eng.NoRoute(noRouteFavicon(), noRouteProxy(&Option{}))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/favicon.ico", nil)
	eng.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/x-icon", w.Header().Get("Content-Type"))
}

func TestDefaultEngineRootFaviconFromStatic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	fsys := fstest.MapFS{"favicon.ico": {Data: []byte("custom-ico")}}
	eng := DefaultEngine(&Option{
		DisabledPprof: true,
		DisableHealth: true,
		DisableGzip:   true,
		StaticFiles:   []*StaticFiles{{FS: fsys, HttpPath: "/"}},
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/favicon.ico", nil)
	eng.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
