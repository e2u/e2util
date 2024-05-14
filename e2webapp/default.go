package e2webapp

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/e2u/e2util/e2conf"
	"github.com/e2u/e2util/e2db"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type DefaultApp struct {
	*e2db.Connect
	*e2conf.Config
	*e2db.Option
}

func NewDefaultApp(cfg *e2conf.Config) *DefaultApp {
	return &DefaultApp{
		Connect: e2db.New(cfg.Orm),
		Config:  cfg,
		Option:  &e2db.Option{Debug: true},
	}
}

type Controller struct {
	*template.Template
}

func NewEmbedFsController(embedTemplates embed.FS, subDir string) *Controller {
	if subDir == "" {
		subDir = "templates"
	}
	templates, _ := fs.Sub(embedTemplates, subDir)
	return NewController(templates)
}

func NewController(templateFS fs.FS) *Controller {
	tmpl := template.New("")
	if err := fs.WalkDir(templateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			data, _ := fs.ReadFile(templateFS, path)
			tmpl, _ = tmpl.New(path).Parse(string(data))
		}
		return nil
	}); err != nil {
		logrus.Fatalf("new controller error=%v", err)
		return nil
	}
	return &Controller{
		Template: tmpl,
	}
}

func AddStaticFs(staticFs fs.FS, r *gin.Engine, httpPath string) {
	r.Use(func(c *gin.Context) {
		if c.Request.URL.Path == httpPath || c.Request.URL.Path == httpPath+"/" {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		c.Next()
	})
	r.StaticFS(httpPath, http.FS(staticFs))
}

func AddEmbedStaticFs(efs embed.FS, r *gin.Engine, subDir string, httpPath string) {
	staticFs, _ := fs.Sub(efs, subDir)
	AddStaticFs(staticFs, r, httpPath)
}
