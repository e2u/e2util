package e2gin

import (
	"html/template"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDynamicHTMLRenderPassesFuncMap(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "hello.html"), []byte(`{{upper "hi"}}`), 0o600))

	r := NewDynamicHTMLRender(dir, template.FuncMap{
		"upper": strings.ToUpper,
	})
	w := httptest.NewRecorder()
	err := r.Instance("hello.html", nil).Render(w)
	require.NoError(t, err)
	assert.Equal(t, "HI", w.Body.String())
}

func TestDynamicHTMLRenderPassesOption(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "space.html"), []byte("<div>\n  hello\n</div>"), 0o600))

	r := NewDynamicHTMLRender(dir, TemplatesOption{MinifyHTML: true})
	w := httptest.NewRecorder()
	err := r.Instance("space.html", nil).Render(w)
	require.NoError(t, err)
	assert.Equal(t, "<div>hello</div>", w.Body.String())
}
