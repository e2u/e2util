package e2app

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
	"log/slog"

	"github.com/e2u/e2util/e2conf"
	"github.com/e2u/e2util/e2db"
)

type FS struct {
	embed.FS
	SubDir string // example: "templates"
}

type Option struct {
	TemplateFs *FS
}

type Application struct {
	*e2db.Connect
	*e2conf.Config
	template *template.Template
}

func (a *Application) ExecuteTemplate(wr io.Writer, name string, data any) error {
	return a.template.ExecuteTemplate(wr, name, data)
}

func (a *Application) Templates() []*template.Template {
	return a.template.Templates()
}

func (a *Application) Template() *template.Template {
	return a.template
}

func New(cfg *e2conf.Config, opt *Option) *Application {
	a := &Application{
		Connect: e2db.New(cfg.Orm),
		Config:  cfg,
		template: func() *template.Template {
			if opt == nil || opt.TemplateFs == nil {
				return nil
			}
			tfs, _ := fs.Sub(opt.TemplateFs.FS, opt.TemplateFs.SubDir)
			tmpl := template.New("")
			if err := fs.WalkDir(tfs, ".", func(path string, d fs.DirEntry, err error) error {
				if !d.IsDir() {
					data, _ := fs.ReadFile(tfs, path)
					tmpl, _ = tmpl.New(path).Parse(string(data))
				}
				return err
			}); err != nil {
				slog.Error("walk template dir error", "error", err)
			}
			return tmpl
		}(),
	}
	return a
}
