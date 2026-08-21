package proxychecker

import (
	"context"
	"net/http"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/e2u/e2util/e2json"
)

func Test_Regest(t *testing.T) {
	reg := regexp.MustCompile(`(?mi)^HTTP/[0-9.]{3}\s[4-5][0-9]{2}\s.+$`)
	cases := []struct {
		line string
		want bool
	}{
		{"HTTP/1.0 200 OK", false},
		{"HTTP/1.0 403 Forbidden", true},
		{"HTTP/1.0 404 Not Found", true},
		{"HTTP/1.0 405 Method Not Allowed", true},
		{"HTTP/1.0 500 Internal Server Error", true},
	}
	for _, tc := range cases {
		if got := reg.MatchString(tc.line); got != tc.want {
			t.Errorf("MatchString(%q) = %v, want %v", tc.line, got, tc.want)
		}
	}
}

func TestDefaultRequestAndCheck(t *testing.T) {
	req := DefaultRequest("https://example.com/").
		WithUserAgent("test-agent").
		WithTimeout(time.Second).
		WithFollowRedirects(false).
		WithHttpMethod(http.MethodHead)
	if err := req.Check(); err != nil {
		t.Fatalf("Check() error: %v", err)
	}
	if req.userAgent != "test-agent" {
		t.Errorf("userAgent = %q", req.userAgent)
	}
	if req.httpMethod != http.MethodHead {
		t.Errorf("httpMethod = %q", req.httpMethod)
	}

	bad := DefaultRequest("ftp://example.com/")
	if err := bad.Check(); err == nil {
		t.Fatal("expected error for unsupported scheme")
	}
}

func TestCheckProxyUnsupportedScheme(t *testing.T) {
	resp := CheckProxy(context.Background(), "socks5://127.0.0.1:1", DefaultRequest("ftp://example.com/"))
	if resp.Error == nil {
		t.Fatal("expected error for unsupported target scheme")
	}
}

func Test_CheckProxy(t *testing.T) {
	proxy := os.Getenv("E2UTIL_PROXY")
	if proxy == "" {
		t.Skip("set E2UTIL_PROXY (e.g. socks5://127.0.0.1:9050) to run live proxy checks")
	}

	ctx := context.Background()
	t.Run("check socks5", func(t *testing.T) {
		resp := CheckProxy(ctx, proxy, DefaultRequest("https://example.com/"))
		if resp.Error != nil {
			t.Fatal(resp.Error)
		}
		t.Log(resp)
	})

	t.Run("check socks5 sina match title", func(t *testing.T) {
		reg := regexp.MustCompile(`<title>手机新浪网</title>`)
		req := DefaultRequest("https://www.sina.cn/").WithExpect(reg)
		resp := CheckProxy(ctx, proxy, req)
		if resp.Error != nil {
			t.Fatal(resp.Error)
		}
		if !resp.MatchExpect {
			t.Fatal(resp)
		}
		t.Log(e2json.MustToJSONString(resp))
	})

	t.Run("check socks5 pincong match title", func(t *testing.T) {
		expectReg := regexp.MustCompile(`<title>Just a moment...</title>`)
		notExpectReg := regexp.MustCompile(`<title>发现 - 新·品葱</title>`)
		req := DefaultRequest("https://pincong.rocks/").WithExpect(expectReg).WithNotExpect(notExpectReg)
		resp := CheckProxy(ctx, proxy, req)
		if resp.Error != nil {
			t.Fatal(resp.Error)
		}
		if !resp.MatchExpect {
			t.Fatal(e2json.MustToJSONString(resp, true))
		}
		if !resp.MatchNotExpect {
			t.Fatal(e2json.MustToJSONString(resp, true))
		}
		t.Log(e2json.MustToJSONString(resp, true))
	})
}
