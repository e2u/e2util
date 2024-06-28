package e2webapp

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"maps"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var startupAt = time.Now()

var (
	FuncMap = make(template.FuncMap)
)

type App interface {
	Routers(p *gin.RouterGroup)
}

func ParseTemplates(templateFs embed.FS) (*template.Template, error) {
	tmpl := template.New("")
	templates, _ := fs.Sub(templateFs, ".")
	defaultFuncMap := template.FuncMap{
		"startAt": func() string { return fmt.Sprintf("v%d", startupAt.Unix()) },
		"add":     func(a, b int) int { return a + b },
		"sub":     func(a, b int) int { return a - b },
		"until":   until,
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
					for i2 := 0; i2 < len(v)-1; i2++ {
						key = v[i2]
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

	if len(FuncMap) > 0 {
		maps.Copy(defaultFuncMap, FuncMap)
	}

	if err := parseTemplates(templates, tmpl, defaultFuncMap); err != nil {
		logrus.Errorf("parset templates error=%v", err)
		return nil, err
	}
	return tmpl, nil
}

func parseTemplates(templateFS fs.FS, tmpl *template.Template, fns template.FuncMap) error {
	if err := fs.WalkDir(templateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			data, _ := fs.ReadFile(templateFS, path)
			tmpl, _ = tmpl.New(path).Funcs(fns).Parse(string(data))
		}
		return nil
	}); err != nil {
		logrus.Fatalf("new controller error=%v", err)
		return err
	}
	return nil
}

// func AddStaticFs(staticFs fs.FS, r *gin.Engine, httpPath string) {
//	httpPath = filepath.Clean(httpPath)
//
//	r.Use(func(c *gin.Context) {
//		if c.Request.URL.Path == httpPath || c.Request.URL.Path == httpPath+"/" {
//			c.AbortWithStatus(http.StatusForbidden)
//			return
//		}
//		c.Next()
//	})
//
//	r.Use(func(c *gin.Context) {
//		if !strings.HasPrefix(c.Request.RequestURI, httpPath) {
//			c.Next()
//			return
//		}
//
//		var etag string
//		fileName := strings.TrimLeft(c.Request.RequestURI, httpPath)
//
//		if before, _, ok := strings.Cut(fileName, "?"); ok {
//			fileName = before
//		}
//
//		if strings.HasSuffix(fileName, "/") {
//			c.AbortWithStatus(http.StatusForbidden)
//			return
//		}
//
//		v, ok := etagCache.Load(fileName)
//		if ok && v != nil {
//			etag = v.(string)
//		} else {
//			b, err := readFileContent(staticFs, fileName)
//			if err != nil {
//				logrus.Errorf("read file error=%v", err)
//				c.Next()
//				return
//			}
//			etag = e2md5.MD5HexString(b)
//			etagCache.Store(fileName, etag)
//		}
//
//		if MatchEtag(c, []byte(etag)) {
//			return
//		}
//
//	}).StaticFS(httpPath, http.FS(staticFs))
//}
//
// func readFileContent(staticFs fs.FS, fileName string) ([]byte, error) {
//	f, err := staticFs.Open(fileName)
//	if err != nil {
//		logrus.Errorf("open file error=%v", err)
//		return nil, err
//	}
//	defer e2exec.MustClose(f)
//	b, err := io.ReadAll(f)
//	if err != nil {
//		logrus.Errorf("read file error=%v", err)
//		return nil, err
//	}
//	return b, nil
//}
//
// func AddEmbedStaticFs(efs embed.FS, r *gin.Engine, httpPath string) {
//	staticFs, _ := fs.Sub(efs, ".")
//	AddStaticFs(staticFs, r, httpPath)
//}
//
// func MatchEtag(c *gin.Context, data []byte) bool {
//	etag := e2md5.HeadTailHex(data)
//	c.Header("Cache-Control", "public, max-age=31536000")
//	c.Header("ETag", etag)
//
//	if match := c.GetHeader("If-None-Match"); match != "" && match == etag {
//		c.Status(http.StatusNotModified)
//		c.Abort()
//		return true
//	}
//	return false
//}
