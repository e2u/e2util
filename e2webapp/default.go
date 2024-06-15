package e2webapp

import (
	"embed"
	"html/template"
	"io/fs"
	"maps"
	"net/http"

	"github.com/e2u/e2util/e2conf"
	"github.com/e2u/e2util/e2db"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type DefaultEnv struct {
	*e2db.Connect
	*e2conf.Config
}

func NewDefaultEnv(cfg *e2conf.Config) *DefaultEnv {
	return &DefaultEnv{
		Connect: e2db.New(cfg.Orm),
		Config:  cfg,
	}
}

type Controller struct {
	*template.Template
}

var FuncMap template.FuncMap

func NewEmbedFsController(embedTemplates embed.FS, subDir string) *Controller {
	if subDir == "" {
		subDir = "templates"
	}
	templates, _ := fs.Sub(embedTemplates, subDir)
	return NewController(templates)
}

func NewController(templateFS fs.FS) *Controller {
	tmpl := template.New("")

	defaultFuncMap := template.FuncMap{
		"add":   func(a, b int) int { return a + b },
		"sub":   func(a, b int) int { return a - b },
		"until": until,
		"trueThen": func(b bool, v any) any {
			if b {
				return v
			}
			return template.HTMLAttr("")
		},
		"falseThen": func(b bool, v any) any {
			if !b {
				return v
			}
			return template.HTMLAttr("")
		},
		"eqThen": func(v1, v2, rv any) any {
			if v1 == v2 {
				return rv
			}
			return template.HTMLAttr("")
		},
		"neThen": func(v1, v2, rv any) any {
			if v1 != v2 {
				return rv
			}
			return template.HTMLAttr("")
		},
		"choose": func(v any, vs map[any]any) any {
			if v, ok := vs[v]; ok {
				return v
			}
			return ""
		},
		"map": func(values ...any) map[string]any {
			if len(values)%2 != 0 {
				return nil
			}
			root := make(map[string]any)
			for i := 0; i < len(values); i += 2 {
				dict := root
				var key string
				switch v := values[i].(type) {
				case string:
					key = v
				case []string:
					for i := 0; i < len(v)-1; i++ {
						key = v[i]
						var m map[string]any
						v, found := dict[key]
						if found {
							m = v.(map[string]any)
						} else {
							m = make(map[string]any)
							dict[key] = m
						}
						dict = m
					}
					key = v[len(v)-1]
				default:
					return nil
				}
				dict[key] = values[i+1]
			}
			return root
		},
	}
	if FuncMap != nil && len(FuncMap) > 0 {
		maps.Copy(defaultFuncMap, FuncMap)
	}

	if err := fs.WalkDir(templateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			data, _ := fs.ReadFile(templateFS, path)
			tmpl, _ = tmpl.New(path).Funcs(defaultFuncMap).Parse(string(data))
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
