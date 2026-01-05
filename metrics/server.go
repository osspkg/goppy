/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package metrics

import (
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.osspkg.com/xc"

	"go.osspkg.com/goppy/v3/env"
	"go.osspkg.com/goppy/v3/web"
)

type Server struct {
	appInfo env.AppInfo
	server  *web.Server
	route   *web.BaseRouter
	conf    Config
}

func New(ctx xc.Context, app env.AppInfo, c Config) *Server {
	router := web.NewBaseRouter()
	conf := web.Config{Addr: c.Addr, Tag: "metric"}
	return &Server{
		server:  web.NewServer(ctx.Context(), conf, router),
		route:   router,
		conf:    c,
		appInfo: app,
	}
}

func (v *Server) Up(ctx xc.Context) error {
	v.pprofRegister()
	v.prometheusRegister()
	return v.server.Up(ctx)
}

func (v *Server) Down() error {
	return v.server.Down()
}

func (v *Server) AddHandler(path string, ctrl func(ctx web.Ctx), methods ...string) {
	v.route.Route(path, ctrl, methods...)
}

func (v *Server) pprofRegister() {
	v.route.Route("/pprof", func(ctx web.Ctx) {
		pprof.Index(ctx.Response(), ctx.Request())
	}, http.MethodGet)
	v.route.Route("/pprof/goroutine", func(ctx web.Ctx) {
		pprof.Index(ctx.Response(), ctx.Request())
	}, http.MethodGet)
	v.route.Route("/pprof/allocs", func(ctx web.Ctx) {
		pprof.Index(ctx.Response(), ctx.Request())
	}, http.MethodGet)
	v.route.Route("/pprof/block", func(ctx web.Ctx) {
		pprof.Index(ctx.Response(), ctx.Request())
	}, http.MethodGet)
	v.route.Route("/pprof/heap", func(ctx web.Ctx) {
		pprof.Index(ctx.Response(), ctx.Request())
	}, http.MethodGet)
	v.route.Route("/pprof/mutex", func(ctx web.Ctx) {
		pprof.Index(ctx.Response(), ctx.Request())
	}, http.MethodGet)
	v.route.Route("/pprof/threadcreate", func(ctx web.Ctx) {
		pprof.Index(ctx.Response(), ctx.Request())
	}, http.MethodGet)
	v.route.Route("/pprof/cmdline", func(ctx web.Ctx) {
		pprof.Cmdline(ctx.Response(), ctx.Request())
	}, http.MethodGet)
	v.route.Route("/pprof/profile", func(ctx web.Ctx) {
		pprof.Profile(ctx.Response(), ctx.Request())
	}, http.MethodGet)
	v.route.Route("/pprof/symbol", func(ctx web.Ctx) {
		pprof.Symbol(ctx.Response(), ctx.Request())
	}, http.MethodGet)
	v.route.Route("/pprof/trace", func(ctx web.Ctx) {
		pprof.Trace(ctx.Response(), ctx.Request())
	}, http.MethodGet)
}

func (v *Server) prometheusRegister() {
	registerAppInfo(v.appInfo)
	registerCounter(string(v.appInfo.AppName), v.conf.Counter)
	registerCounterVec(string(v.appInfo.AppName), v.conf.CounterVec)
	registerGauge(string(v.appInfo.AppName), v.conf.Gauge)
	registerGaugeVec(string(v.appInfo.AppName), v.conf.GaugeVec)
	registerHistogram(string(v.appInfo.AppName), v.conf.Histogram)
	registerHistogramVec(string(v.appInfo.AppName), v.conf.HistogramVec)

	handler := promhttp.HandlerFor(object.prometheus, promhttp.HandlerOpts{Registry: object.prometheus})
	v.route.Route("/metrics", func(ctx web.Ctx) {
		handler.ServeHTTP(ctx.Response(), ctx.Request())
	}, http.MethodGet)
}
