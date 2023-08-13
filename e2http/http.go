package e2http

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/e2u/e2util/e2json"
)

type Context struct {
	cli            *http.Client
	ctx            context.Context
	url            *url.URL
	method         string
	reqHeaders     http.Header
	req            *http.Request
	reqBody        io.Reader
	respBody       []byte
	respHeaders    http.Header
	delHeaders     []string
	respStatusCode int
	errs           []error
	toJsonPointer  any
	outWriter      io.Writer
	dumpReqWriter  io.Writer
	dumpBody       bool
}

func Builder(ctx context.Context) *Context {
	return &Context{
		cli:        &http.Client{},
		ctx:        ctx,
		method:     http.MethodGet,
		reqHeaders: make(map[string][]string),
	}
}

func (r *Context) URL(u string) *Context {
	if v, err := url.Parse(u); err == nil {
		r.url = v
	} else {
		r.appendErr(err)
	}
	return r
}

// Proxy set proxy, e.g. socks5://127.0.0.1:1080, http://127.0.0.1:3128
func (r *Context) Proxy(p string) *Context {
	proxyUrl, err := url.Parse(p)
	if err != nil {
		r.appendErr(err)
		return r
	}
	r.cli.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	return r
}

func (r *Context) DumpRequest(w io.Writer, body bool) *Context {
	if r.req != nil {
		if b, err := httputil.DumpRequestOut(r.req, body); err == nil {
			_, _ = io.Copy(w, bytes.NewReader(b))
		} else {
			r.appendErr(err)
		}
	} else {
		r.dumpReqWriter = w
		r.dumpBody = body
	}
	return r
}

func (r *Context) Method(m string) {
	r.method = m
}

func (r *Context) Post(rd io.Reader) *Context {
	return r.postForm(http.MethodPost, rd)
}

func (r *Context) Put(rd io.Reader) *Context {
	return r.postForm(http.MethodPut, rd)
}

func (r *Context) postForm(method string, rd io.Reader) *Context {
	r.Method(method)
	r.ContentType("application/x-www-form-urlencoded")
	r.reqBody = rd
	return r
}

func (r *Context) PostMultipart(values map[string]io.Reader) *Context {
	return r.postMultipart(http.MethodPost, values)
}

func (r *Context) PutMultipart(values map[string]io.Reader) *Context {
	return r.postMultipart(http.MethodPut, values)
}

func (r *Context) postMultipart(method string, values map[string]io.Reader) *Context {
	r.Method(method)
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	cloz := func(x io.Reader) {
		if rd, ok := x.(io.Closer); ok && rd != nil {
			_ = rd.Close()
		}
	}
	for key, rd := range values {
		var fw io.Writer
		var err error
		if x, ok := rd.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				cloz(rd)
				break
			}
		} else {
			if fw, err = w.CreateFormField(key); err != nil {
				cloz(rd)
				break
			}
		}
		if _, err = io.Copy(fw, rd); err != nil {
			cloz(rd)
			break
		}
		cloz(rd)
	}

	_ = w.Close()
	r.ContentType(w.FormDataContentType())
	r.reqBody = &buf
	return r
}

func (r *Context) SetHeaders(h map[string]string) *Context {
	for k, v := range h {
		r.reqHeaders.Set(k, v)
	}
	return r
}

func (r *Context) UserAgent(u string) *Context {
	r.reqHeaders.Set("User-Agent", u)
	return r
}

func (r *Context) ContentType(c string) *Context {
	r.reqHeaders.Set("Content-Type", c)
	return r
}

func (r *Context) AddHeader(key, value string) *Context {
	r.reqHeaders.Add(key, value)
	return r
}

func (r *Context) SetHeader(key, value string) *Context {
	r.reqHeaders.Set(key, value)
	return r
}

func (r *Context) RemoveHeader(key string) *Context {
	r.delHeaders = append(r.delHeaders, key)
	return r
}

func (r *Context) ToJSON(t any) *Context {
	if len(r.respBody) != 0 {
		if err := e2json.MustFromJSONByte(r.respBody, t); err != nil {
			r.appendErr(err)
		}
	} else {
		r.toJsonPointer = t
	}
	return r
}

func (r *Context) Write(w io.Writer) *Context {
	if len(r.respBody) != 0 {
		_, _ = io.Copy(w, bytes.NewReader(r.respBody))
	} else {
		r.outWriter = w
	}
	return r
}

func (r *Context) Do() *Context {

	if req, err := http.NewRequestWithContext(r.ctx, r.method, r.url.String(), r.reqBody); err == nil {
		r.req = req
	} else {
		r.appendErr(err)
		return r
	}

	r.req.Header = r.reqHeaders

	for _, k := range r.delHeaders {
		r.req.Header.Del(k)
	}

	resp, err := r.cli.Do(r.req)
	if err != nil {
		r.appendErr(err)
		return r
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	r.respStatusCode = resp.StatusCode
	r.respHeaders = resp.Header

	if b, err := io.ReadAll(resp.Body); err == nil {
		r.respBody = b
	} else {
		r.appendErr(err)
		return r
	}

	if r.toJsonPointer != nil {
		if err := e2json.MustFromJSONByte(r.respBody, r.toJsonPointer); err != nil {
			r.appendErr(err)
			return r
		}
	}

	if r.outWriter != nil {
		if _, err := io.Copy(r.outWriter, bytes.NewReader(r.respBody)); err != nil {
			r.appendErr(err)
			return r
		}
	}

	if r.dumpReqWriter != nil {
		if b, err := httputil.DumpRequestOut(r.req, r.dumpBody); err == nil {
			if _, err := io.Copy(r.dumpReqWriter, bytes.NewReader(b)); err != nil {
				r.appendErr(err)
				return r
			}
		} else {
			r.appendErr(err)
			return r
		}
	}
	return r
}

func (r *Context) StatusCode() int {
	return r.respStatusCode
}

func (r *Context) Headers() http.Header {
	return r.respHeaders
}

func (r *Context) Body() []byte {
	return r.respBody
}

func (r *Context) BodyString() string {
	return string(r.respBody)
}

func (r *Context) appendErr(err error) {
	r.errs = append(r.errs, err)
}

func (r *Context) Errors() []error {
	return r.errs
}
