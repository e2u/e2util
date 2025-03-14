package e2gin

import (
	"html/template"
	"os"
	"sync"
	"time"

	"github.com/e2u/e2util/e2io"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin/render"
	"github.com/sirupsen/logrus"
)

type DynamicHTMLRender struct {
	mu sync.RWMutex
	// tu          sync.Mutex
	templates   *template.Template
	eventTimers map[string]*time.Timer
	dir         string
	args        any
}

func NewDynamicHTMLRender(dir string, args ...any) *DynamicHTMLRender {
	dhr := &DynamicHTMLRender{
		dir:         dir,
		args:        args,
		eventTimers: make(map[string]*time.Timer),
	}
	dhr.reloadTemplates()
	go dhr.watchTemplates()
	return dhr
}

func (d *DynamicHTMLRender) Instance(name string, data interface{}) render.Render {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return &render.HTML{
		Template: d.templates,
		Name:     name,
		Data:     data,
	}
}

func (d *DynamicHTMLRender) reloadTemplates() {
	d.mu.Lock()
	defer d.mu.Unlock()

	tmpl, err := ParseTemplates(os.DirFS(d.dir), d.args)
	if err != nil {
		logrus.Errorf("Failed to reload templates: %v", err)
		return
	}
	d.templates = tmpl
	logrus.Info("Templates reloaded successfully.")
}

func (d *DynamicHTMLRender) watchTemplates() {
	e2io.WatchDir(d.dir, func(s string, event fsnotify.Event) {
		logrus.Infof("Template changed: %s", event.Name)
		d.reloadTemplates()
	})
}
