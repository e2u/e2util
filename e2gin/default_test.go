package e2gin

import (
	"io/fs"
	"os"
	"testing"
)

var content = `<html>
{{if}}
{{end}}

{{ if }}
{{ end }}

{{          if          }}
{{          end              }}

{{	if	}}
{{	end	}}

{{-if-}}
{{-end-}}

{{- if -}}
{{- end -}}
</html>`

func Test_trimTags(t *testing.T) {
	t.Log(trimTags(content))
}

func Test_trimTags_minifyHTML(t *testing.T) {
	t.Log(minifyHTML(trimTags(content)))
}

func Test_fs(t *testing.T) {
	tmpDir := os.DirFS("/tmp/")
	f, err := fs.Sub(tmpDir, ".")
	if err != nil {
		t.Fatal(err)
	}

	err = fs.WalkDir(f, ".", func(path string, d fs.DirEntry, err error) error {
		t.Log(path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
