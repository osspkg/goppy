package http

//go:generate easyjson

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	nethttp "net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/deweppro/go-http/pkg/httputil"
	"github.com/deweppro/go-http/pkg/httputil/dec"
	"github.com/deweppro/go-http/pkg/httputil/enc"
	"github.com/deweppro/go-http/pkg/routes"
	"github.com/deweppro/go-http/servers"
	"github.com/deweppro/go-http/servers/web"
	"github.com/deweppro/go-logger"
)

type (
	ctx struct {
		w nethttp.ResponseWriter
		r *nethttp.Request
		l logger.Logger
	}

	//Ctx request and response interface
	Ctx interface {
		URL() *url.URL
		Redirect(uri string)
		Param(key string) Paramer
		GetHead(key string) string
		SetHead(key, value string)
		GetCookie(key string) *nethttp.Cookie
		SetCookie(value *nethttp.Cookie)
		GetBody() BodyReader
		SetBody() BodyWriter
		Context() context.Context
		Log() logger.LogWriter
	}
)

func newCtx(w nethttp.ResponseWriter, r *nethttp.Request, l logger.Logger) *ctx {
	return &ctx{
		w: w,
		r: r,
		l: l,
	}
}

type (
	//Paramer interface for typing a parameter from a URL
	Paramer interface {
		String() (string, error)
		Int() (int64, error)
		Float() (float64, error)
	}
	param struct {
		val string
		err error
	}
)

//String getting the parameter as a string
func (v param) String() (string, error) { return v.val, v.err }

//Int getting the parameter as a int64
func (v param) Int() (int64, error) {
	if v.err != nil {
		return 0, v.err
	}
	return strconv.ParseInt(v.val, 10, 64)
}

//Float getting the parameter as a float64
func (v param) Float() (float64, error) {
	if v.err != nil {
		return 0.0, v.err
	}
	return strconv.ParseFloat(v.val, 64)
}

//Param getting a parameter from URL by key
func (v *ctx) Param(key string) Paramer {
	val, err := httputil.VarsString(v.r, key)
	return param{
		val: val,
		err: err,
	}
}

//Log log entry interface
func (v *ctx) Log() logger.LogWriter {
	return v.l
}

//GetHead getting headers from a key request
func (v *ctx) GetHead(key string) string {
	return v.r.Header.Get(key)
}

//SetHead setting response headers
func (v *ctx) SetHead(key, value string) {
	v.w.Header().Set(key, value)
}

//GetCookie getting cookies from a key request
func (v *ctx) GetCookie(key string) *nethttp.Cookie {
	c, _ := v.r.Cookie(key) //nolint: errcheck
	return c
}

//SetCookie setting cookies in response
func (v *ctx) SetCookie(value *nethttp.Cookie) {
	nethttp.SetCookie(v.w, value)
}

type (
	//BodyReader request body reading interface
	BodyReader interface {
		Raw() []byte
		JSON(in interface{}) error
	}

	bodyReader struct {
		r *nethttp.Request
	}
)

//Raw getting the raw request body
func (v *bodyReader) Raw() []byte {
	defer v.r.Body.Close() //nolint: errcheck
	b, _ := ioutil.ReadAll(v.r.Body)
	return b
}

//JSON decoding the request body into a structure
func (v *bodyReader) JSON(in interface{}) error { return dec.JSON(v.r, in) }

//GetBody request body handler
func (v *ctx) GetBody() BodyReader {
	return &bodyReader{r: v.r}
}

type (
	//BodyWriter response body record interface
	BodyWriter interface {
		JSON(in interface{})
		Stream(in []byte, filename string)
		Raw(in []byte)
		Error(in ErrMessage)
	}

	//ErrMessage standard error message type
	//easyjson:json
	ErrMessage struct {
		HTTPCode     int                    `json:"-"`
		InternalCode string                 `json:"code"`
		Message      string                 `json:"msg"`
		Ctx          map[string]interface{} `json:"ctx,omitempty"`
	}

	bodyWriter struct {
		w nethttp.ResponseWriter
	}
)

//Raw recording the response in raw format
func (v *bodyWriter) Raw(in []byte) { enc.Raw(v.w, in) }

//JSON recording the response in json format
func (v *bodyWriter) JSON(in interface{}) { enc.JSON(v.w, in) }

//Stream sending raw data in response with the definition of the content type by the file name
func (v *bodyWriter) Stream(in []byte, filename string) { enc.Stream(v.w, in, filename) }

//Error recording an error response
func (v *bodyWriter) Error(in ErrMessage) {
	b, _ := json.Marshal(&in) //nolint: errcheck
	v.w.Header().Add("Content-Type", "application/json; charset=utf-8")
	v.w.WriteHeader(in.HTTPCode)
	v.w.Write(b) //nolint: errcheck
}

//SetBody response body handler
func (v *ctx) SetBody() BodyWriter {
	return &bodyWriter{w: v.w}
}

//Context provider the request context
func (v *ctx) Context() context.Context {
	return v.r.Context()
}

//URL getting a URL from a request
func (v *ctx) URL() *url.URL {
	uri := v.r.URL
	uri.Host = v.r.Host
	uri.Scheme = v.r.Proto
	return uri
}

//Redirect redirecting to another URL
func (v *ctx) Redirect(uri string) {
	nethttp.Redirect(v.w, v.r, uri, nethttp.StatusMovedPermanently)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	routePoolItem struct {
		active bool
		route  *route
	}
	//RouterPool router pool handler
	RouterPool struct {
		pool map[string]*routePoolItem
	}
)

func newRoutePool(configs map[string]servers.Config, log logger.Logger) *RouterPool {
	v := &RouterPool{
		pool: make(map[string]*routePoolItem),
	}
	for name, config := range configs {
		v.pool[name] = &routePoolItem{
			active: false,
			route:  newRouter(config, log),
		}
	}
	return v
}

//All method to get all route handlers
func (v *RouterPool) All(call func(name string, router Router)) {
	for n, r := range v.pool {
		call(n, r.route)
	}
}

//Main method to get Main route handler
func (v *RouterPool) Main() Router {
	return v.Get("main")
}

//Get method to get route handler by key
func (v *RouterPool) Get(name string) Router {
	if r, ok := v.pool[name]; ok {
		return r.route
	}
	panic(fmt.Sprintf("Route with name `%s` is not found", name))
}

func (v *RouterPool) Up() error {
	for n, r := range v.pool {
		r.active = true
		if err := r.route.Up(); err != nil {
			return fmt.Errorf("pool `%s`: %w", n, err)
		}
	}
	return nil
}

func (v *RouterPool) Down() error {
	for n, r := range v.pool {
		if !r.active {
			continue
		}
		if err := r.route.Down(); err != nil {
			return fmt.Errorf("pool `%s`: %w", n, err)
		}
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type (
	route struct {
		r  *routes.Router
		ws *web.Server
		c  servers.Config
		l  logger.Logger
	}

	//Middleware type of middleware
	Middleware func(func(nethttp.ResponseWriter, *nethttp.Request)) func(nethttp.ResponseWriter, *nethttp.Request)

	//Router router handler interface
	Router interface {
		Use(args ...Middleware)
		Collection(prefix string, args ...Middleware) RouteCollector

		RouteCollector
	}
)

func newRouter(conf servers.Config, log logger.Logger) *route {
	return &route{
		r: routes.NewRouter(),
		c: conf,
		l: log,
	}
}

func (v *route) Up() error {
	v.ws = web.New(v.c, v.r, v.l)
	return v.ws.Up()
}
func (v *route) Down() error {
	return v.ws.Down()
}

func (v *route) Use(args ...Middleware) {
	for _, arg := range args {
		arg := arg
		v.r.Global(func(ctrlFunc routes.CtrlFunc) routes.CtrlFunc {
			return arg(ctrlFunc)
		})
	}
}

func (v *route) Match(path string, call func(ctx Ctx), methods ...string) {
	v.r.Route(path, func(w nethttp.ResponseWriter, r *nethttp.Request) {
		call(newCtx(w, r, v.l))
	}, methods...)
}

func (v *route) Get(path string, call func(ctx Ctx))     { v.Match(path, call, nethttp.MethodGet) }
func (v *route) Head(path string, call func(ctx Ctx))    { v.Match(path, call, nethttp.MethodHead) }
func (v *route) Post(path string, call func(ctx Ctx))    { v.Match(path, call, nethttp.MethodPost) }
func (v *route) Put(path string, call func(ctx Ctx))     { v.Match(path, call, nethttp.MethodPut) }
func (v *route) Delete(path string, call func(ctx Ctx))  { v.Match(path, call, nethttp.MethodDelete) }
func (v *route) Options(path string, call func(ctx Ctx)) { v.Match(path, call, nethttp.MethodOptions) }
func (v *route) Patch(path string, call func(ctx Ctx))   { v.Match(path, call, nethttp.MethodPatch) }

type (
	//RouteCollector interface of the router collection
	RouteCollector interface {
		Get(path string, call func(ctx Ctx))
		Head(path string, call func(ctx Ctx))
		Post(path string, call func(ctx Ctx))
		Put(path string, call func(ctx Ctx))
		Delete(path string, call func(ctx Ctx))
		Options(path string, call func(ctx Ctx))
		Patch(path string, call func(ctx Ctx))
		Match(path string, call func(ctx Ctx), methods ...string)
	}

	rc struct {
		prefix string
		route  *route
	}
)

func (v *rc) Match(path string, call func(ctx Ctx), methods ...string) {
	path = strings.TrimLeft(path, "/")
	v.route.Match(v.prefix+"/"+path, call, methods...)
}

func (v *rc) Get(path string, call func(ctx Ctx))     { v.Match(path, call, nethttp.MethodGet) }
func (v *rc) Head(path string, call func(ctx Ctx))    { v.Match(path, call, nethttp.MethodHead) }
func (v *rc) Post(path string, call func(ctx Ctx))    { v.Match(path, call, nethttp.MethodPost) }
func (v *rc) Put(path string, call func(ctx Ctx))     { v.Match(path, call, nethttp.MethodPut) }
func (v *rc) Delete(path string, call func(ctx Ctx))  { v.Match(path, call, nethttp.MethodDelete) }
func (v *rc) Options(path string, call func(ctx Ctx)) { v.Match(path, call, nethttp.MethodOptions) }
func (v *rc) Patch(path string, call func(ctx Ctx))   { v.Match(path, call, nethttp.MethodPatch) }

//Collection route collection handler
func (v *route) Collection(prefix string, args ...Middleware) RouteCollector {
	prefix = strings.TrimRight(prefix, "/")
	for _, arg := range args {
		arg := arg
		v.r.Middlewares(prefix, func(ctrlFunc routes.CtrlFunc) routes.CtrlFunc {
			return arg(ctrlFunc)
		})
	}

	return &rc{
		prefix: prefix,
		route:  v,
	}
}
