package proxychecker

import (
	"context"
	"regexp"
	"testing"

	"github.com/e2u/e2util/e2json"
)

func Test_Regest(t *testing.T) {

	t.Run("t1", func(t *testing.T) {
		str200 := "HTTP/1.0 200 OK"

		str403 := "HTTP/1.0 403 Forbidden"
		str404 := "HTTP/1.0 404 Not Found"
		str405 := "HTTP/1.0 405 Method Not Allowed"

		str500 := "HTTP/1.0 500 Internal Server Error"

		reg := regexp.MustCompile(`(?mi)^HTTP/[0-9.]{3}\s[4-5][0-9]{2}\s.+$`)
		t.Log(str200, reg.MatchString(str200))
		t.Log(str403, reg.MatchString(str403))
		t.Log(str404, reg.MatchString(str404))
		t.Log(str405, reg.MatchString(str405))
		t.Log(str500, reg.MatchString(str500))

	})
}

func Test_CheckProxy(t *testing.T) {
	ctx := context.Background()
	t.Run("check socks5", func(t *testing.T) {
		// https://ifconfig.io/all.json
		location := ""
		location = "https://ifconfig.io/all.json"
		location = "https://www.163.com/"
		// location = "https://www.google.com/"
		resp := CheckProxy(ctx,
			"socks5://192.168.10.29:9050",
			DefaultRequest(location),
		)
		if resp.Error != nil {
			t.Fatal(resp.Error)
		}
		t.Log(resp)
	})

	t.Run("check socks5 sina match title", func(t *testing.T) {
		// https://ifconfig.io/all.json
		location := ""
		location = "https://ifconfig.io/all.json"
		// location = "https://www.163.com/"
		location = "https://www.sina.cn/"
		// location = "https://www.google.com/"
		reg := regexp.MustCompile(`<title>手机新浪网</title>`)
		req := DefaultRequest(location).WithExpect(reg) //.WithDumpResponse(os.Stdout)
		resp := CheckProxy(ctx,
			"socks5://192.168.10.29:9050",
			req,
		)
		if resp.Error != nil {
			t.Fatal(resp.Error)
		}
		if !resp.MatchExpect {
			t.Fatal(resp)
		}
		t.Log(e2json.MustToJSONString(resp))
	})

	t.Run("check socks5 pincong match title", func(t *testing.T) {
		// https://ifconfig.io/all.json
		location := ""
		location = "https://ifconfig.io/all.json"
		location = "https://pincong.rocks/"
		expectReg := regexp.MustCompile(`<title>Just a moment...</title>`)
		notExpectReg := regexp.MustCompile(`<title>发现 - 新·品葱</title>`)
		req := DefaultRequest(location).WithExpect(expectReg).WithNotExpect(notExpectReg) //.WithDumpResponse(os.Stdout)
		resp := CheckProxy(ctx,
			"socks5://192.168.10.29:9050",
			req,
		)
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
