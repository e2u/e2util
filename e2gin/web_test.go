package e2gin

import (
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStaticFS implements fs.FS for testing
type mockStaticFS struct {
	files map[string]string
}

func (m *mockStaticFS) Open(name string) (fs.File, error) {
	if content, ok := m.files[name]; ok {
		return &mockFile{content: content, name: name}, nil
	}
	return nil, os.ErrNotExist
}

func (m *mockStaticFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return nil, nil
}

type mockFile struct {
	content string
	name    string
	offset  int
}

func (f *mockFile) Stat() (fs.FileInfo, error) { return nil, nil }
func (f *mockFile) Read(b []byte) (int, error) {
	if f.offset >= len(f.content) {
		return 0, io.EOF
	}
	n := copy(b, f.content[f.offset:])
	f.offset += n
	return n, nil
}
func (f *mockFile) Close() error { return nil }

// Helper to create test StaticFiles
type testFileSystem struct {
	indexContent   string
	loginContent   string
	aboutContent   string
	hasIndex       bool
	hasLogin       bool
	hasAbout       bool
}

func createTestStaticFiles(tfs testFileSystem) []*StaticFiles {
	files := make(map[string]string)
	if tfs.hasIndex {
		files["index.html"] = tfs.indexContent
	}
	if tfs.hasLogin {
		files["login.html"] = tfs.loginContent
	}
	if tfs.hasAbout {
		files["about.html"] = tfs.aboutContent
	}

	return []*StaticFiles{
		{
			FS:       &mockStaticFS{files: files},
			HttpPath: "/",
		},
	}
}

func setupTestRouter(sfs []*StaticFiles) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Register the NoRoute handler
	r.NoRoute(noRouteStaticIndex(sfs))

	return r
}

// TestSPA_Routing tests SPA routing behavior
func TestSPA_Routing(t *testing.T) {
	tests := []struct {
		name           string
		staticFiles    testFileSystem
		path           string
		wantStatus     int
		wantBody       string
		wantBodyPrefix string
		description    string
	}{
		{
			name: "SPA route returns index.html",
			staticFiles: testFileSystem{
				indexContent: "<html>SPA Index</html>",
				hasIndex:     true,
				hasLogin:     false,
			},
			path:        "/login",
			wantStatus:  http.StatusOK,
			wantBody:    "<html>SPA Index</html>",
			description: "When login.html doesn't exist, should return index.html for SPA",
		},
		{
			name: "SPA route returns index.html for dashboard",
			staticFiles: testFileSystem{
				indexContent: "<html>SPA Index</html>",
				hasIndex:     true,
			},
			path:        "/dashboard/users/123",
			wantStatus:  http.StatusOK,
			wantBody:    "<html>SPA Index</html>",
			description: "Deep nested paths should return index.html for SPA routing",
		},
		{
			name: "Non-SPA route returns specific page",
			staticFiles: testFileSystem{
				indexContent: "<html>SPA Index</html>",
				loginContent: "<html>Login Page</html>",
				hasIndex:     true,
				hasLogin:     true,
			},
			path:        "/login",
			wantStatus:  http.StatusOK,
			wantBody:    "<html>Login Page</html>",
			description: "When login.html exists, should return login.html instead of index.html",
		},
		{
			name: "Non-SPA about page",
			staticFiles: testFileSystem{
				indexContent: "<html>SPA Index</html>",
				aboutContent: "<html>About Page</html>",
				hasIndex:     true,
				hasAbout:     true,
			},
			path:        "/about",
			wantStatus:  http.StatusOK,
			wantBody:    "<html>About Page</html>",
			description: "When about.html exists, should return about.html",
		},
		{
			name: "Root path returns index",
			staticFiles: testFileSystem{
				indexContent: "<html>SPA Index</html>",
				hasIndex:     true,
			},
			path:        "/",
			wantStatus:  http.StatusOK,
			wantBody:    "<html>SPA Index</html>",
			description: "Root path should return index.html",
		},
		{
			name: "Explicit HTML extension",
			staticFiles: testFileSystem{
				loginContent: "<html>Login Page</html>",
				hasLogin:     true,
			},
			path:        "/login.html",
			wantStatus:  http.StatusOK,
			wantBody:    "<html>Login Page</html>",
			description: "Explicit .html extension should work",
		},
		{
			name: "No index.html returns 404",
			staticFiles: testFileSystem{
				hasIndex: false,
				hasLogin: false,
			},
			path:        "/some-path",
			wantStatus:  http.StatusNotFound,
			description: "When no index.html and no matching page, should 404",
		},
		{
			name: "Non-browser request skips HTML handling",
			staticFiles: testFileSystem{
				indexContent: "<html>SPA Index</html>",
				hasIndex:     true,
			},
			path:        "/api/data",
			wantStatus:  http.StatusNotFound,
			description: "API paths should skip HTML handling",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sfs := createTestStaticFiles(tt.staticFiles)
			router := setupTestRouter(sfs)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			req.Header.Set("Accept", "text/html")

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code, tt.description)
			if tt.wantBody != "" {
				assert.Equal(t, tt.wantBody, w.Body.String(), tt.description)
			}
			if tt.wantBodyPrefix != "" {
				assert.True(t, len(w.Body.String()) > 0 && w.Body.String()[:len(tt.wantBodyPrefix)] == tt.wantBodyPrefix)
			}
		})
	}
}

// TestSPA_AcceptHeader tests that only browser requests get HTML
func TestSPA_AcceptHeader(t *testing.T) {
	sfs := createTestStaticFiles(testFileSystem{
		indexContent: "<html>SPA Index</html>",
		hasIndex:     true,
	})
	router := setupTestRouter(sfs)

	tests := []struct {
		name       string
		accept     string
		wantStatus int
	}{
		{
			name:       "text/html accept",
			accept:     "text/html",
			wantStatus: http.StatusOK,
		},
		{
			name:       "*/* accept",
			accept:     "*/*",
			wantStatus: http.StatusOK,
		},
		{
			name:       "empty accept",
			accept:     "",
			wantStatus: http.StatusOK,
		},
		{
			name:       "application/json accept",
			accept:     "application/json",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "text/css accept",
			accept:     "text/css",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/dashboard", nil)
			if tt.accept != "" {
				req.Header.Set("Accept", tt.accept)
			}

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// TestSPA_Security tests path traversal prevention
func TestSPA_Security(t *testing.T) {
	sfs := createTestStaticFiles(testFileSystem{
		indexContent: "<html>SPA Index</html>",
		hasIndex:     true,
	})
	router := setupTestRouter(sfs)

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "path traversal with dotdot",
			path:       "/../etc/passwd",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "path traversal encoded",
			path:       "/%2e%2e/etc/passwd",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "path traversal double dotdot",
			path:       "/../../etc/passwd",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "normal path with dots",
			path:       "/v1.0/page",
			wantStatus: http.StatusOK, // Should return index.html (SPA)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			req.Header.Set("Accept", "text/html")

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// TestSPA_HTTPMethods tests that only GET and HEAD are handled
func TestSPA_HTTPMethods(t *testing.T) {
	sfs := createTestStaticFiles(testFileSystem{
		indexContent: "<html>SPA Index</html>",
		hasIndex:     true,
	})
	router := setupTestRouter(sfs)

	tests := []struct {
		method     string
		wantStatus int
	}{
		{"GET", http.StatusOK},
		{"HEAD", http.StatusOK},
		{"POST", http.StatusNotFound},
		{"PUT", http.StatusNotFound},
		{"DELETE", http.StatusNotFound},
		{"PATCH", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, "/login", nil)
			req.Header.Set("Accept", "text/html")

			router.ServeHTTP(w, req)
			assert.Equal(t, tt.wantStatus, w.Code, "Method %s should return %d", tt.method, tt.wantStatus)
		})
	}
}

// TestSPA_MultipleStaticFiles tests with multiple StaticFiles configurations
func TestSPA_MultipleStaticFiles(t *testing.T) {
	// Create multiple static file sources
	files1 := map[string]string{
		"index.html": "<html>App1 Index</html>",
	}
	files2 := map[string]string{
		"login.html": "<html>App1 Login</html>",
	}

	sfs := []*StaticFiles{
		{
			FS:       &mockStaticFS{files: files1},
			HttpPath: "/",
		},
		{
			FS:       &mockStaticFS{files: files2},
			HttpPath: "/",
		},
	}

	router := setupTestRouter(sfs)

	t.Run("finds page from second StaticFiles", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/login", nil)
		req.Header.Set("Accept", "text/html")

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "<html>App1 Login</html>", w.Body.String())
	})

	t.Run("finds index from first StaticFiles", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("Accept", "text/html")

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "<html>App1 Index</html>", w.Body.String())
	})
}

// TestSPA_ContentType tests Content-Type header
func TestSPA_ContentType(t *testing.T) {
	sfs := createTestStaticFiles(testFileSystem{
		indexContent: "<html>SPA Index</html>",
		hasIndex:     true,
	})
	router := setupTestRouter(sfs)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("Accept", "text/html")

	router.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	assert.Equal(t, "text/html; charset=utf-8", contentType)
}

// TestLoadHTMLPage tests the loadHTMLPage function
func TestLoadHTMLPage(t *testing.T) {
	t.Run("loads existing page", func(t *testing.T) {
		sfs := createTestStaticFiles(testFileSystem{
			loginContent: "<html>Login</html>",
			hasLogin:     true,
		})

		content := loadHTMLPage(sfs, "login")
		require.NotNil(t, content)
		assert.Equal(t, "<html>Login</html>", string(content))
	})

	t.Run("returns nil for non-existent page", func(t *testing.T) {
		sfs := createTestStaticFiles(testFileSystem{
			hasIndex: true,
		})

		content := loadHTMLPage(sfs, "nonexistent")
		assert.Nil(t, content)
	})
}

// TestHasHTMLPage tests the hasHTMLPage function
func TestHasHTMLPage(t *testing.T) {
	t.Run("returns true for existing page", func(t *testing.T) {
		sfs := createTestStaticFiles(testFileSystem{
			loginContent: "<html>Login</html>",
			hasLogin:     true,
		})

		assert.True(t, hasHTMLPage(sfs, "login"))
	})

	t.Run("returns false for non-existent page", func(t *testing.T) {
		sfs := createTestStaticFiles(testFileSystem{
			hasIndex: true,
		})

		assert.False(t, hasHTMLPage(sfs, "nonexistent"))
	})
}

// TestIsSafePath tests the isSafePath function
func TestIsSafePath(t *testing.T) {
	tests := []struct {
		path    string
		wantSafe bool
	}{
		{"login.html", true},
		{"pages/about.html", true},
		{"v1.0/app.js", true},
		{"../etc/passwd", false},
		{"../../config.json", false},
		{"login/../../../etc/passwd", false},
		{"page?name=test", false}, // query strings should not be in path
		{"", false},
		{"/absolute/path", true}, // leading slash is allowed (will be normalized)
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := isSafePath(tt.path)
			assert.Equal(t, tt.wantSafe, got, "Path %s safety check", tt.path)
		})
	}
}

// TestGetContentType tests the getContentType function
func TestGetContentType(t *testing.T) {
	tests := []struct {
		ext      string
		wantType string
	}{
		{"/style.css", "text/css; charset=utf-8"},
		{"/app.js", "application/javascript; charset=utf-8"},
		{"/index.html", "text/html; charset=utf-8"},
		{"/data.json", "application/json"},
		{"/image.png", "image/png"},
		{"/icon.ico", "image/x-icon"},
		{"/font.woff2", "font/woff2"},
		{"/unknown.xyz", "application/octet-stream"},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			got := getContentType(tt.ext)
			assert.Equal(t, tt.wantType, got)
		})
	}
}
