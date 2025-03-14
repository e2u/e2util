package e2gin

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"testing"
	"time"

	"github.com/e2u/e2util/e2exec"
	"github.com/gin-contrib/cors"
)

var (
	webFS          embed.FS
	embedAssets    embed.FS
	embedTemplates embed.FS
)

func TestDefaultEngine(t *testing.T) {
	tf := template.FuncMap{
		"baseUrl": func() template.URL { return template.URL("baseUrl") },
		"url": func(path string) template.URL {
			return template.URL(fmt.Sprintf("%s%s", "baseUrl", path))
		},
	}
	r := DefaultEngine(&Option{
		DisableGzip: false,
		StaticFiles: []*StaticFiles{
			{
				FS:       e2exec.Must(fs.Sub(webFS, "mycash-web/build")),
				HttpPath: "/",
			},
			{
				FS:       embedAssets,
				HttpPath: "/assets",
			},
		},
		//HTMLTemplate: e2exec.Must(ParseTemplates(embedTemplates, tf)),
		Template: &Template{
			FS:      embedTemplates,
			FuncMap: tf,
			Option:  TemplatesOption{},
		},
	})

	r.Use(cors.New(cors.Config{
		AllowAllOrigins:           true,
		AllowMethods:              []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders:              []string{"Origin", "Content-Type", "Authorization", "Range", "X-Api-Consumer"},
		ExposeHeaders:             []string{"Content-Range", "X-Total-Count"},
		AllowCredentials:          true,
		MaxAge:                    12 * time.Hour,
		AllowWebSockets:           true,
		AllowFiles:                true,
		OptionsResponseStatusCode: 200,
	}))

	apiGroup := r.Group("/api/v1")
	_ = apiGroup
}
