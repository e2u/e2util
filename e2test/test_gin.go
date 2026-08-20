package e2test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/e2u/e2util/e2json"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func PostJSON(r *gin.Engine, url string, header http.Header, body any) *httptest.ResponseRecorder {
	return Any(r, http.MethodPost, url, header, body)
}

func PutJSON(r *gin.Engine, url string, header http.Header, body any) *httptest.ResponseRecorder {
	return Any(r, http.MethodPut, url, header, body)
}

func Get(r *gin.Engine, url string, header http.Header, params url.Values) *httptest.ResponseRecorder {
	if params != nil && params.Encode() != "" {
		url = url + "?" + params.Encode()
	}
	return Any(r, http.MethodGet, url, header, nil)
}

func Any(r *gin.Engine, method string, url string, header http.Header, body any) *httptest.ResponseRecorder {
	logrus.Infof("e2test [%s] %s", method, url)

	ctx := context.TODO()

	var rd io.Reader
	if body != nil {
		switch t := body.(type) {
		case io.Reader:
			rd = t
		case string:
			rd = strings.NewReader(t)
		case []byte:
			rd = bytes.NewReader(t)
		default:
			rd = strings.NewReader(e2json.MustToJSONString(body))
		}
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequestWithContext(ctx, method, url, rd)
	if header != nil {
		req.Header = header
	}
	r.ServeHTTP(w, req)
	return w
}
