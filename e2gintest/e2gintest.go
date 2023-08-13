package e2gintest

import (
	"io"
	"maps"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
)

/**

func InitTestSuitConfig() *e2conf.Config {
	return e2conf.New(&e2conf.InitConfigInput{
		Env:           "test",
		AddConfigPath: []string{"../../etc/"},
		ConfigName:    "app-dev",
	})
}

func InitTestSuitApp() *e2app.Application {
	app := e2app.New(InitTestSuitConfig(), &e2app.Option{})
	return app
}


func Test_DownloadMergedFile(t *testing.T) {
	a := app.InitTestSuitApp()
	tr := e2gintest.New(a)
	rs := tr.Run(&e2gintest.Request{
		RegUri:  "/project/download/:project_id",
		ReqUri:  "/project/download/90d0b1f2-b091-48a0-9aba-a1bcf257036b",
		Handler: NewProject(a).DownloadMergedFile,
		Method:  http.MethodGet,
	})
	t.Log(rs.Header())
	t.Log(rs.Result())

}
*/

type Request struct {
	RegUri   string
	ReqUri   string
	Method   string
	Body     io.Reader
	Header   http.Header
	Handlers []gin.HandlerFunc
}

type Gin struct {
	engine *gin.Engine
}

func New() *Gin {
	return &Gin{
		engine: gin.Default(),
	}
}

type CloseNotifierResponseRecorder struct {
	*httptest.ResponseRecorder
	closeNotifyChan chan bool
}

func (c *CloseNotifierResponseRecorder) CloseNotify() <-chan bool {
	return c.closeNotifyChan
}

func (g *Gin) Run(r *Request) *CloseNotifierResponseRecorder {
	var req *http.Request
	w := &CloseNotifierResponseRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		closeNotifyChan:  make(chan bool),
	}

	if r.Method == "" {
		r.Method = http.MethodGet
	}
	if r.RegUri == "" {
		r.RegUri = r.ReqUri
	}

	g.engine.Handle(r.Method, r.RegUri, r.Handlers...)
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		req, _ = http.NewRequest(r.Method, r.ReqUri, r.Body)
	case http.MethodGet, http.MethodOptions:
		req, _ = http.NewRequest(r.Method, r.ReqUri, nil)
	}

	maps.Copy(req.Header, r.Header)
	g.engine.ServeHTTP(w, req)
	return w
}
