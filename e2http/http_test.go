package e2http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuilderGetHtml(t *testing.T) {
	var buf bytes.Buffer
	r := Builder(context.TODO()).
		URL("https://www.apache.org").
		Write(&buf).
		Do()

	assert.Equal(t, len(r.Errors()), 0)
	assert.Equal(t, r.StatusCode(), http.StatusOK)
	assert.Contains(t, r.BodyString(), "<title>Welcome to The Apache Software Foundation!</title>")
}

func TestBuilderJson(t *testing.T) {
	type St struct {
		IP        string `json:"ip"`
		Port      int    `json:"port"`
		Ping      int    `json:"ping"`
		Protocols []struct {
			Type     string `json:"type"`
			Port     int    `json:"port"`
			TLS      bool   `json:"tls"`
			AutoRead []bool `json:"autoRead"`
		} `json:"protocols"`
	}

	var j []St

	var buf bytes.Buffer
	var dump bytes.Buffer

	f, _ := os.Open("/tmp/1.js")

	pm := map[string]io.Reader{
		"aaa": strings.NewReader("hello"),
		"ccc": strings.NewReader("123456789"),
		"bbb": f,
	}
	_ = pm

	r := Builder(context.TODO()).
		Proxy("http://127.0.0.1:8888").
		URL("https://raw.githubusercontent.com/jetkai/proxy-list/main/online-proxies/json/proxies-basic.json").
		Write(&buf).
		ToJSON(&j).
		DumpRequest(&dump, true).
		AddHeader("aaaa", "1").
		UserAgent("http-client/0.1").
		RemoveHeader("Accept-Encoding").
		RemoveHeader("AAAA").
		Do()

	fmt.Println(dump.String())
	assert.Equal(t, len(r.Errors()), 0)
	assert.Equal(t, r.StatusCode(), http.StatusOK)
	assert.Contains(t, r.BodyString(), "https")
	assert.Contains(t, r.BodyString(), "http")
	assert.Contains(t, r.BodyString(), "8080")
	assert.Greater(t, len(j), 0)

}
