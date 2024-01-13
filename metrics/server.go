/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package metrics

import (
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.osspkg.com/goppy/env"
	"go.osspkg.com/goppy/web"
	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xlog"
)

type Server struct {
	app    env.AppInfo
	server *web.Server
	route  *web.BaseRouter
	conf   Config
}

func New(app env.AppInfo, c Config, l xlog.Logger) *Server {
	router := web.NewBaseRouter()
	conf := web.Config{Addr: c.Addr}
	return &Server{
		server: web.NewServer("Metrics", conf, router, l),
		route:  router,
		conf:   c,
		app:    app,
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

func (v *Server) AddHandler(path string, ctrl func(http.ResponseWriter, *http.Request), methods ...string) {
	v.route.Route(path, ctrl, methods...)
}

func (v *Server) pprofRegister() {
	v.route.Route("/pprof", pprof.Index, http.MethodGet)
	v.route.Route("/pprof/goroutine", pprof.Index, http.MethodGet)
	v.route.Route("/pprof/allocs", pprof.Index, http.MethodGet)
	v.route.Route("/pprof/block", pprof.Index, http.MethodGet)
	v.route.Route("/pprof/heap", pprof.Index, http.MethodGet)
	v.route.Route("/pprof/mutex", pprof.Index, http.MethodGet)
	v.route.Route("/pprof/threadcreate", pprof.Index, http.MethodGet)
	v.route.Route("/pprof/cmdline", pprof.Cmdline, http.MethodGet)
	v.route.Route("/pprof/profile", pprof.Profile, http.MethodGet)
	v.route.Route("/pprof/symbol", pprof.Symbol, http.MethodGet)
	v.route.Route("/pprof/trace", pprof.Trace, http.MethodGet)
}

func (v *Server) prometheusRegister() {
	registerAppInfo(v.app)
	registerCounter(string(v.app.AppName), v.conf.Counter)
	registerCounterVec(string(v.app.AppName), v.conf.CounterVec)
	registerGauge(string(v.app.AppName), v.conf.Gauge)
	registerGaugeVec(string(v.app.AppName), v.conf.GaugeVec)
	registerHistogram(string(v.app.AppName), v.conf.Histogram)
	registerHistogramVec(string(v.app.AppName), v.conf.HistogramVec)

	handler := promhttp.HandlerFor(object.prometheus, promhttp.HandlerOpts{Registry: object.prometheus})
	v.route.Route("/metrics", func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}, http.MethodGet)
}
