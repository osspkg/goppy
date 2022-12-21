package http

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/dewep-online/goppy/internal"
	"github.com/dewep-online/goppy/plugins"
	"github.com/deweppro/go-http/pkg/signature"
)

type (
	HTTPClientConfig struct {
		Config *HTTPClientConfigItem `yaml:"httpcli"`
	}

	HTTPClientConfigItem struct {
		Proxy               string        `yaml:"proxy"`
		Timeout             time.Duration `yaml:"timeout"`
		KeepAlive           time.Duration `yaml:"keepalive"`
		MaxIdleConns        int           `yaml:"maxidleconns"`
		MaxIdleConnsPerHost int           `yaml:"maxidleconnsperhost"`
	}
)

func (c *HTTPClientConfig) Default() {
	if c.Config == nil {
		c.Config = &HTTPClientConfigItem{
			Proxy:               "",
			Timeout:             5 * time.Second,
			KeepAlive:           60 * time.Second,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// WithHTTPClient init pool http clients
func WithHTTPClient() plugins.Plugin {
	return plugins.Plugin{
		Config: &HTTPClientConfig{},
		Inject: func(conf *HTTPClientConfig) HTTPClient {
			return newHTTPClient(conf.Config)
		},
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	httpCli struct {
		cli *http.Client
	}

	HTTPClient interface {
		Create(call func(rb HTTPRequestBind)) HTTPResponseBind
	}
)

func createProxy(proxy string) func(r *http.Request) (*url.URL, error) {
	if len(proxy) == 0 {
		return http.ProxyFromEnvironment
	}
	u, err := url.Parse(proxy)
	if err != nil {
		return func(r *http.Request) (*url.URL, error) {
			return nil, err
		}
	}
	return http.ProxyURL(u)
}

func newHTTPClient(c *HTTPClientConfigItem) HTTPClient {
	return &httpCli{
		cli: &http.Client{
			Transport: &http.Transport{
				Proxy: createProxy(c.Proxy),
				DialContext: (&net.Dialer{
					Timeout:   c.KeepAlive,
					KeepAlive: c.KeepAlive,
				}).DialContext,
				MaxIdleConns:        c.MaxIdleConns,
				MaxIdleConnsPerHost: c.MaxIdleConnsPerHost,
			},
			Timeout: c.Timeout,
		},
	}
}

func (c *httpCli) Create(call func(rb HTTPRequestBind)) HTTPResponseBind {
	req := newHTTPClientRequest()
	call(req)

	if req.e != nil {
		return newHTTPClientResponse(0, nil, nil, req.e)
	}

	hreq, err := http.NewRequest(req.m, req.u, bytes.NewReader(req.b))
	if err != nil {
		return newHTTPClientResponse(0, nil, nil, err)
	}

	req.Header("Connection", "keep-alive")
	for k := range req.h {
		hreq.Header.Set(k, req.h.Get(k))
	}

	if req.s != nil {
		signature.Encode(hreq.Header, req.s, req.b)
	}

	hres, err := c.cli.Do(hreq)
	if err != nil {
		return newHTTPClientResponse(0, nil, nil, err)
	}

	b, err := internal.ReadAll(hres.Body)
	return newHTTPClientResponse(hres.StatusCode, b, hres.Header.Clone(), err)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	httpClientRequest struct {
		m string
		u string
		b []byte
		h http.Header
		s Signature
		e error
	}

	HTTPRequestBind interface {
		Method(v string)
		URI(v string)
		Body(v []byte)
		Header(k, v string)
		Signature(v Signature)
		JSON(v interface{})
	}

	httpClientResponse struct {
		c int
		b []byte
		h http.Header
		e error
	}

	HTTPResponseBind interface {
		Code() int
		Body() []byte
		Headers() http.Header
		Err() error
		JSON(v interface{}) error
	}
)

func newHTTPClientRequest() *httpClientRequest {
	return &httpClientRequest{
		m: http.MethodGet,
		h: make(http.Header),
	}
}
func (c *httpClientRequest) Method(v string) { c.m = v }
func (c *httpClientRequest) URI(v string)    { c.u = v }
func (c *httpClientRequest) Body(v []byte) {
	if len(c.b) > 0 {
		return
	}
	c.b = v
}
func (c *httpClientRequest) Header(k, v string)    { c.h.Set(k, v) }
func (c *httpClientRequest) Signature(v Signature) { c.s = v }
func (c *httpClientRequest) JSON(v interface{}) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.b, c.e = json.Marshal(v)
}

func newHTTPClientResponse(c int, b []byte, h http.Header, e error) *httpClientResponse {
	return &httpClientResponse{
		c: c,
		b: b,
		h: h,
		e: e,
	}
}
func (c *httpClientResponse) Code() int            { return c.c }
func (c *httpClientResponse) Body() []byte         { return c.b }
func (c *httpClientResponse) Headers() http.Header { return c.h }
func (c *httpClientResponse) Err() error           { return c.e }
func (c *httpClientResponse) JSON(v interface{}) error {
	if c.e != nil {
		return c.e
	}
	return json.Unmarshal(c.b, v)
}
