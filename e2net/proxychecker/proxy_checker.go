package proxychecker

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/e2u/e2util/e2exec"
	"github.com/e2u/e2util/e2http"
)

const (
	DefaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"
	DefaultTimeout   = 30 * time.Second
)

var (
	DefaultExpect    = regexp.MustCompile("(?mi)^HTTP/[0-9.]{3} 200 OK$")
	DefaultNotExpect = regexp.MustCompile(`(?mi)^HTTP/[0-9.]{3}\s[4-5][0-9]{2}\s.+$`)
	DefaultHeader    = http.Header{
		"User-Agent": []string{DefaultUserAgent},
		"Accept":     []string{"*/*"},
	}
)

// User-Agents
// https://github.com/monperrus/crawler-user-agents/blob/master/crawler-user-agents.json

type Request struct {
	Url                *url.URL
	userAgent          string
	headers            http.Header
	expect             *regexp.Regexp
	notExpect          *regexp.Regexp
	timeout            time.Duration
	followRedirects    bool
	httpMethod         string
	dumpResponseWriter io.Writer
}

type Response struct {
	Status         int
	Proxy          string
	Target         string
	Duration       time.Duration
	MatchExpect    bool
	MatchNotExpect bool
	Error          error
}

func (r *Request) Check() error {
	if r.Url.Scheme != "http" && r.Url.Scheme != "https" {
		return errors.New("unsupported protocol scheme")
	}
	if r.timeout <= 0 {
		r.timeout = DefaultTimeout
	}
	if r.userAgent == "" {
		r.userAgent = DefaultUserAgent
	}
	if r.httpMethod == "" {
		r.httpMethod = http.MethodGet
	}
	if r.headers == nil {
		r.headers = DefaultHeader
	}
	return nil
}

func DefaultRequest(location string) *Request {
	return &Request{
		Url:             e2exec.Must(url.Parse(location)),
		userAgent:       DefaultUserAgent,
		headers:         DefaultHeader,
		expect:          DefaultExpect,
		timeout:         DefaultTimeout,
		followRedirects: true,
		httpMethod:      http.MethodGet,
	}
}

func (r *Request) WithExpect(reg *regexp.Regexp) *Request {
	r.expect = reg
	return r
}

func (r *Request) WithNotExpect(reg *regexp.Regexp) *Request {
	r.notExpect = reg
	return r
}

func (r *Request) WithUserAgent(agent string) *Request {
	r.userAgent = agent
	return r
}

func (r *Request) WithDumpResponse(w io.Writer) *Request {
	r.dumpResponseWriter = w
	return r
}

func (r *Request) WithTimeout(timeout time.Duration) *Request {
	r.timeout = timeout
	return r
}

func (r *Request) WithFollowRedirects(followRedirects bool) *Request {
	r.followRedirects = followRedirects
	return r
}

func (r *Request) WithHttpMethod(method string) *Request {
	r.httpMethod = method
	return r
}

func (r *Request) CombineHeaders(headers http.Header) *Request {
	if r.headers == nil {
		r.headers = http.Header{}
	}
	for k, v := range headers {
		if len(v) > 0 {
			r.headers.Add(k, v[0])
		}
	}
	return r
}
func (r *Request) ReplaceHeaders(headers http.Header) *Request {
	if headers != nil {
		r.headers = headers
	}
	return r
}

func CheckProxy(ctx context.Context, proxy string, target *Request) *Response {
	resp := &Response{
		Proxy:  proxy,
		Target: target.Url.String(),
	}

	if err := target.Check(); err != nil {
		resp.Error = err
		return resp
	}

	var buf bytes.Buffer
	start := time.Now()
	h := e2http.Builder(ctx).
		Proxy(proxy).
		ConnectTimeout(target.timeout).
		UserAgent(target.userAgent).
		URL(target.Url.String()).
		DumpResponse(&buf, true)
	errs := h.Do().Errors()
	resp.Duration = time.Since(start)

	if len(errs) > 0 {
		resp.Error = errors.Join(errs...)
		return resp
	}

	if target.dumpResponseWriter != nil {
		if _, err := io.Copy(target.dumpResponseWriter, &buf); err != nil {
			resp.Error = err
			return resp
		}
	}

	resp.Status = h.StatusCode()
	if target.expect == nil && target.notExpect == nil {
		return resp
	}

	scanner := bufio.NewScanner(&buf)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if target.expect != nil && target.expect.MatchString(line) {
			resp.MatchExpect = true
		}
		if target.notExpect != nil && !target.notExpect.MatchString(line) {
			resp.MatchNotExpect = true
		}
	}

	return resp
}
