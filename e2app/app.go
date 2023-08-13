package e2app

import (
	"embed"
	"html/template"
	"io"
	"log/slog"

	"github.com/e2u/e2util/e2conf"
	"github.com/e2u/e2util/e2db"
	"github.com/e2u/e2util/e2template"
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
			tmpl, err := e2template.FromEmbedFS(&opt.TemplateFs.FS, opt.TemplateFs.SubDir)
			if err != nil {
				slog.Error("parse templates", "error", err)
			}
			return tmpl
		}(),
	}
	return a
}
