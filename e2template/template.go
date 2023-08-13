package e2template

import (
	"embed"
	"errors"
	"html/template"
	"io/fs"
	"log/slog"
)

func FromEmbedFS(efs *embed.FS, subDir string) (*template.Template, error) {
	if efs == nil {
		return nil, errors.New("template embed fs nil")
	}
	tfs, _ := fs.Sub(efs, subDir)
	tmpl := template.New("")
	if err := fs.WalkDir(tfs, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			data, _ := fs.ReadFile(tfs, path)
			tmpl, _ = tmpl.New(path).Parse(string(data))
		}
		return err
	}); err != nil {
		slog.Error("walk template dir error", "error", err)
		return nil, err
	}
	return tmpl, nil
}
