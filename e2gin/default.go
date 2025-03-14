package e2gin

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"maps"
	"regexp"
	"strings"
	"time"

	"github.com/e2u/e2util/e2crypto"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

var startupAt = time.Now()

var (
	FuncMap = make(template.FuncMap)
)

type App[T any] interface {
	Routers(r *gin.RouterGroup) T
}

type TemplatesOption struct {
	TrimTags   bool // if TrimTags was true, all tags {{ and }} of the template would add - as {{- and -}}
	MinifyHTML bool
}

var defaultFuncMap = template.FuncMap{
	"nonce": func() string {
		return e2crypto.RandomString(16)
	},
	"startAt": func() string {
		if gin.IsDebugging() {
			return fmt.Sprintf("v%d", time.Now().Unix())
		}
		return fmt.Sprintf("v%d", startupAt.Unix())
	},
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
	"trim": strings.TrimSpace,
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

func ParseTemplates(templateFs fs.FS, args ...any) (*template.Template, error) {
	tmpl := template.New("")
	templates, _ := fs.Sub(templateFs, ".")

	opt := TemplatesOption{}
	for _, arg := range args {
		if v, ok := arg.(template.FuncMap); ok && len(v) > 0 {
			maps.Copy(defaultFuncMap, v)
		}
		if v, ok := arg.(TemplatesOption); ok {
			opt = v
		}
	}

	if len(FuncMap) > 0 {
		maps.Copy(defaultFuncMap, FuncMap)
	}

	if err := parseTemplates(templates, tmpl, defaultFuncMap, opt); err != nil {
		logrus.Errorf("parset templates error=%v", err)
		return nil, err
	}
	return tmpl, nil
}

func parseTemplates(templateFS fs.FS, tmpl *template.Template, fns template.FuncMap, option TemplatesOption) error {
	if err := fs.WalkDir(templateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		defer func() {
			if err := recover(); err != nil {
				logrus.WithField("type", "recover").Errorf("parse templates error=%v, path=%v", err, path)
			}
		}()

		dataByte, _ := fs.ReadFile(templateFS, path)
		dataString := string(dataByte)
		if option.TrimTags {
			dataString = trimTags(dataString)
		}
		if option.MinifyHTML {
			dataString = minifyHTML(dataString)
		}

		if _, err := tmpl.New(path).Funcs(fns).Parse(dataString); err != nil {
			_, _ = tmpl.New(path).Funcs(fns).Parse(errorPage("parse template error", err))
		}

		return nil
	}); err != nil {
		logrus.Errorf("parse templates error=%v", err)
		return err
	}
	return nil
}

var trimTagRegexS = regexp.MustCompile(`{{\s*([^\-])`)
var trimTagRegexE = regexp.MustCompile(`([^\-])\s*}}`)

func trimTags(content string) string {
	content = trimTagRegexS.ReplaceAllString(content, "{{- $1")
	content = trimTagRegexE.ReplaceAllString(content, "$1 -}}")
	return content
}

func minifyHTML(input string) string {
	var buf bytes.Buffer
	tokenizer := html.NewTokenizer(strings.NewReader(input))

	skipMinify := false

	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return buf.String()

		case html.StartTagToken:
			name, _ := tokenizer.TagName()
			tag := string(name)

			if tag == "pre" || tag == "textarea" {
				skipMinify = true
			}

			buf.Write(tokenizer.Raw())

		case html.EndTagToken:
			name, _ := tokenizer.TagName()
			tag := string(name)

			if tag == "pre" || tag == "textarea" {
				skipMinify = false
			}

			buf.Write(tokenizer.Raw())

		case html.TextToken:
			text := string(tokenizer.Text())
			if skipMinify {
				buf.WriteString(text)
			} else {
				minified := strings.TrimSpace(text)
				if minified != "" {
					buf.WriteString(minified)
				}
			}

		default:
			buf.Write(tokenizer.Raw())
		}
	}
}
