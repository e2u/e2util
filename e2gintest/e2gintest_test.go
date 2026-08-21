package e2gintest

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRunGet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	g := New()
	w := g.Run(&Request{
		ReqUri: "/ping",
		Method: http.MethodGet,
		Handlers: []gin.HandlerFunc{
			func(c *gin.Context) { c.String(http.StatusOK, "pong") },
		},
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if body := w.Body.String(); body != "pong" {
		t.Fatalf("body = %q, want pong", body)
	}
}

func TestRunPostAndHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)
	g := New()
	w := g.Run(&Request{
		RegUri: "/echo",
		ReqUri: "/echo",
		Method: http.MethodPost,
		Body:   strings.NewReader(`{"ok":true}`),
		Header: http.Header{"Content-Type": []string{"application/json"}, "X-Test": []string{"1"}},
		Handlers: []gin.HandlerFunc{
			func(c *gin.Context) {
				if c.GetHeader("X-Test") != "1" {
					c.String(http.StatusBadRequest, "missing header")
					return
				}
				c.String(http.StatusCreated, "ok")
			},
		},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", w.Code)
	}
	if w.CloseNotify() == nil {
		t.Fatal("CloseNotify channel is nil")
	}
}
