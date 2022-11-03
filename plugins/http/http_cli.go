package http

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
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
	ClientConfig struct {
		Config *ClientConfigItem `yaml:"httpcli"`
	}

	ClientConfigItem struct {
		Proxy               string        `yaml:"proxy"`
		Timeout             time.Duration `yaml:"timeout"`
		KeepAlive           time.Duration `yaml:"keepalive"`
		MaxIdleConns        int           `yaml:"maxidleconns"`
		MaxIdleConnsPerHost int           `yaml:"maxidleconnsperhost"`
	}
)

func (c *ClientConfig) Default() {
	if c.Config == nil {
		c.Config = &ClientConfigItem{
			Proxy:               "",
			Timeout:             5 * time.Second,
			KeepAlive:           60 * time.Second,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		}
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// WithHTTPClients init pool http clients
func WithHTTPClients() plugins.Plugin {
	return plugins.Plugin{
		Config: &ClientConfig{},
		Inject: func(conf *ClientConfig) Client {
			return newClient(conf.Config)
		},
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	cli struct {
		cli *http.Client
	}

	Client interface {
		Create(call func(rb RequestBind)) ResponseBind
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

func newClient(c *ClientConfigItem) Client {
	return &cli{
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

func (c *cli) Create(call func(rb RequestBind)) ResponseBind {
	req := newCliReq()
	call(req)

	if req.e != nil {
		return newCliRes(0, nil, nil, req.e)
	}

	hreq, err := http.NewRequest(req.m, req.u, bytes.NewReader(req.b))
	if err != nil {
		return newCliRes(0, nil, nil, err)
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
		return newCliRes(0, nil, nil, err)
	}

	b, err := internal.ReadAll(hres.Body)
	return newCliRes(hres.StatusCode, b, hres.Header.Clone(), err)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	cliReq struct {
		m string
		u string
		b []byte
		h http.Header
		s Signature
		e error
	}

	RequestBind interface {
		Method(v string)
		URI(v string)
		Body(v []byte)
		Header(k, v string)
		Signature(v Signature)
		JSON(v interface{})
	}

	cliRes struct {
		c int
		b []byte
		h http.Header
		e error
	}

	ResponseBind interface {
		Code() int
		Body() []byte
		Headers() http.Header
		Err() error
		JSON(v interface{}) error
	}
)

func newCliReq() *cliReq {
	return &cliReq{
		m: http.MethodGet,
		h: make(http.Header),
	}
}
func (c *cliReq) Method(v string) { c.m = v }
func (c *cliReq) URI(v string)    { c.u = v }
func (c *cliReq) Body(v []byte) {
	if len(c.b) > 0 {
		return
	}
	c.b = v
}
func (c *cliReq) Header(k, v string)    { c.h.Set(k, v) }
func (c *cliReq) Signature(v Signature) { c.s = v }
func (c *cliReq) JSON(v interface{}) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.b, c.e = json.Marshal(v)
}

func newCliRes(c int, b []byte, h http.Header, e error) *cliRes {
	return &cliRes{
		c: c,
		b: b,
		h: h,
		e: e,
	}
}
func (c *cliRes) Code() int            { return c.c }
func (c *cliRes) Body() []byte         { return c.b }
func (c *cliRes) Headers() http.Header { return c.h }
func (c *cliRes) Err() error           { return c.e }
func (c *cliRes) JSON(v interface{}) error {
	if c.e != nil {
		return c.e
	}
	return json.Unmarshal(c.b, v)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type Signature interface {
	ID() string
	Algorithm() string
	Create(b []byte) []byte
	CreateString(b []byte) string
	Validate(b []byte, ex string) bool
}

// NewSHA1 create sign sha1
func NewSHA1(id, secret string) Signature {
	return signature.NewCustomSignature(id, secret, "hmac-sha1", sha1.New)
}

// NewSHA256 create sign sha256
func NewSHA256(id, secret string) Signature {
	return signature.NewCustomSignature(id, secret, "hmac-sha256", sha256.New)
}

// NewSHA512 create sign sha512
func NewSHA512(id, secret string) Signature {
	return signature.NewCustomSignature(id, secret, "hmac-sha512", sha512.New)
}
