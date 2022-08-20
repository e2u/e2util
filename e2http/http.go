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

type Request struct {
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

func Builder(ctx context.Context) *Request {
	return &Request{
		cli:        http.DefaultClient,
		ctx:        ctx,
		method:     http.MethodGet,
		reqHeaders: make(map[string][]string),
	}
}

func (r *Request) URL(u string) *Request {
	if v, err := url.Parse(u); err == nil {
		r.url = v
	} else {
		r.appendErr(err)
	}
	return r
}

// Proxy set proxy, e.g. socks5://127.0.0.1:1080, http://127.0.0.1:3128
func (r *Request) Proxy(p string) *Request {
	proxyUrl, err := url.Parse(p)
	if err != nil {
		r.appendErr(err)
		return r
	}
	http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	return r
}

func (r *Request) DumpRequest(w io.Writer, body bool) *Request {
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

func (r *Request) Method(m string) {
	r.method = m
}

func (r *Request) Post(rd io.Reader) *Request {
	return r.postForm(http.MethodPost, rd)
}

func (r *Request) Put(rd io.Reader) *Request {
	return r.postForm(http.MethodPut, rd)
}

func (r *Request) postForm(method string, rd io.Reader) *Request {
	r.Method(method)
	r.ContentType("application/x-www-form-urlencoded")
	r.reqBody = rd
	return r
}

func (r *Request) PostMultipart(values map[string]io.Reader) *Request {
	return r.postMultipart(http.MethodPost, values)
}

func (r *Request) PutMultipart(values map[string]io.Reader) *Request {
	return r.postMultipart(http.MethodPut, values)
}

func (r *Request) postMultipart(method string, values map[string]io.Reader) *Request {
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

func (r *Request) SetHeaders(h map[string]string) *Request {
	for k, v := range h {
		r.reqHeaders.Set(k, v)
	}
	return r
}

func (r *Request) UserAgent(u string) *Request {
	r.reqHeaders.Set("User-Agent", u)
	return r
}

func (r *Request) ContentType(c string) *Request {
	r.reqHeaders.Set("Content-Type", c)
	return r
}

func (r *Request) AddHeader(key, value string) *Request {
	r.reqHeaders.Add(key, value)
	return r
}

func (r *Request) SetHeader(key, value string) *Request {
	r.reqHeaders.Set(key, value)
	return r
}

func (r *Request) RemoveHeader(key string) *Request {
	r.delHeaders = append(r.delHeaders, key)
	return r
}

func (r *Request) ToJSON(t any) *Request {
	if len(r.respBody) != 0 {
		if err := e2json.MustFromJSONByte(r.respBody, t); err != nil {
			r.appendErr(err)
		}
	} else {
		r.toJsonPointer = t
	}
	return r
}

func (r *Request) Write(w io.Writer) *Request {
	if len(r.respBody) != 0 {
		_, _ = io.Copy(w, bytes.NewReader(r.respBody))
	} else {
		r.outWriter = w
	}
	return r
}

func (r *Request) Run() *Request {
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
		}
	}

	if r.outWriter != nil {
		_, _ = io.Copy(r.outWriter, bytes.NewReader(r.respBody))
	}

	if r.dumpReqWriter != nil {
		if b, err := httputil.DumpRequestOut(r.req, r.dumpBody); err == nil {
			_, _ = io.Copy(r.dumpReqWriter, bytes.NewReader(b))
		} else {
			r.appendErr(err)
		}
	}
	return r
}

func (r *Request) StatusCode() int {
	return r.respStatusCode
}

func (r *Request) Headers() http.Header {
	return r.respHeaders
}

func (r *Request) Body() []byte {
	return r.respBody
}

func (r *Request) BodyString() string {
	return string(r.respBody)
}

func (r *Request) Errors() []error {
	return r.errs
}

func (r *Request) appendErr(err error) {
	r.errs = append(r.errs, err)
}
