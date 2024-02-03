package e2template

import (
	"embed"
	"errors"
	"html/template"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func FromEmbedFS(efs *embed.FS, subDir string, fcs template.FuncMap) (*template.Template, error) {
	if efs == nil {
		return nil, errors.New("template embed fs nil")
	}
	tfs, _ := fs.Sub(efs, subDir)
	tmpl := template.New("")
	if fcs != nil && len(fcs) > 0 {
		tmpl.Funcs(fcs)
	}
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

func FromPath(templatePath string, fcs template.FuncMap) (*template.Template, error) {
	if templatePath == "" {
		return nil, errors.New("template path not set")
	}
	tmpl := template.New("")

	if fcs != nil && len(fcs) > 0 {
		tmpl.Funcs(fcs)
	}

	if !strings.HasSuffix(templatePath, "/") {
		templatePath += "/"
	}
	if err := filepath.WalkDir(templatePath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && path != "" {
			data, _ := os.ReadFile(path)
			tmpl, _ = tmpl.New(strings.Replace(path, templatePath, "", 1)).Parse(string(data))
		}
		return err
	}); err != nil {
		slog.Error("walk template dir error", "error", err)
		return nil, err
	}
	return tmpl, nil
}
