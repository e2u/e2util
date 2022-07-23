package e2gintest

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/e2u/e2util/e2json"
	"github.com/gin-gonic/gin"
)

/**
func testSetup(t *testing.T) (*gin.Engine, *handler.Handler) {
	config := e2conf.New(&e2conf.InitConfigInput{Env: "dev", ConfigFs: cfgFS})
	const root = "/product"
	router := e2gin.DefaultEngine(&e2gin.Option{
		Root:          root,
		DisabledPprof: true,
	})
	hand := handler.New(config)
	return router, hand
}

func TestGenCode(t *testing.T) {
	r, h := testSetup(t)

	u := "/gen-code"
	var output interface{}
	wr := e2gintest.RunGet(u, r, h.GenCode, &output)
	assert.Equal(t, http.StatusOK, wr.Code)
}
*/

func RunTest(url, method string, router *gin.Engine, handler gin.HandlerFunc, input any, output any) *httptest.ResponseRecorder {
	var req *http.Request

	switch method {
	case http.MethodPost, http.MethodPut:
		router.POST(url, handler)
		req, _ = http.NewRequest(method, url, bytes.NewReader(e2json.MustToJSONByte(input)))
	case http.MethodGet:
		router.GET(url, handler)
		req, _ = http.NewRequest(method, url, nil)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	respRaw, _ := ioutil.ReadAll(w.Body)
	_ = e2json.MustFromJSONByte(respRaw, output)
	return w
}

func RenPost(url string, router *gin.Engine, handler gin.HandlerFunc, input any, output any) *httptest.ResponseRecorder {
	return RunTest(url, http.MethodPost, router, handler, input, output)
}

func RunGet(url string, router *gin.Engine, handler gin.HandlerFunc, output any) *httptest.ResponseRecorder {
	return RunTest(url, http.MethodGet, router, handler, nil, output)
}
