/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package metrics

import (
	"runtime"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.osspkg.com/goppy/env"
)

var (
	once   sync.Once
	object *prom
)

type (
	prom struct {
		prometheus   *prometheus.Registry
		counter      map[string]prometheus.Counter
		counterVec   map[string]*prometheus.CounterVec
		gauge        map[string]prometheus.Gauge
		gaugeVec     map[string]*prometheus.GaugeVec
		histogram    map[string]prometheus.Histogram
		histogramVec map[string]*prometheus.HistogramVec
	}
)

func init() {
	once.Do(func() {
		object = &prom{
			prometheus:   prometheus.NewRegistry(),
			counter:      make(map[string]prometheus.Counter, 2),
			counterVec:   make(map[string]*prometheus.CounterVec, 2),
			gauge:        make(map[string]prometheus.Gauge, 2),
			gaugeVec:     make(map[string]*prometheus.GaugeVec, 2),
			histogram:    make(map[string]prometheus.Histogram, 2),
			histogramVec: make(map[string]*prometheus.HistogramVec, 2),
		}
	})
}

const (
	labelVersion = "version"
	labelOS      = "os"
	labelArch    = "arch"
)

func registerAppInfo(app env.AppInfo) {
	desc := prometheus.NewDesc(
		string(app.AppName)+"_build_info",
		string(app.AppDescription),
		[]string{labelVersion, labelOS, labelArch},
		nil)
	v := prometheus.NewMetricVec(desc, func(lvs ...string) prometheus.Metric {
		return appInfo{desc: desc, labelPairs: prometheus.MakeLabelPairs(desc, lvs)}
	})
	if _, err := v.GetMetricWithLabelValues(string(app.AppVersion), runtime.GOOS, runtime.GOARCH); err != nil {
		panic(err)
	}
	object.prometheus.MustRegister(v)
}

func registerCounter(app string, opts []string) {
	for _, name := range opts {
		if _, ok := object.counter[name]; ok {
			continue
		}
		v := prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: app,
			Name:      name,
		})
		object.counter[name] = v
		object.prometheus.MustRegister(v)
	}
}

func registerCounterVec(app string, opts map[string][]string) {
	for name, labels := range opts {
		if _, ok := object.counterVec[name]; ok {
			continue
		}
		v := prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: app,
			Name:      name,
		}, labels)
		object.counterVec[name] = v
		object.prometheus.MustRegister(v)
	}
}

func registerGauge(app string, opts []string) {
	for _, name := range opts {
		if _, ok := object.gauge[name]; ok {
			continue
		}
		v := prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: app,
			Name:      name,
		})
		object.gauge[name] = v
		object.prometheus.MustRegister(v)
	}
}

func registerGaugeVec(app string, opts map[string][]string) {
	for name, labels := range opts {
		if _, ok := object.gaugeVec[name]; ok {
			continue
		}
		v := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: app,
			Name:      name,
		}, labels)
		object.gaugeVec[name] = v
		object.prometheus.MustRegister(v)
	}
}

func registerHistogram(app string, opts map[string][]float64) {
	for name, buckets := range opts {
		if _, ok := object.histogram[name]; ok {
			continue
		}
		v := prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: app,
			Name:      name,
			Buckets:   buckets,
		})
		object.histogram[name] = v
		object.prometheus.MustRegister(v)
	}
}
func registerHistogramVec(app string, opts map[string]Buckets) {
	for name, opt := range opts {
		if _, ok := object.histogramVec[name]; ok {
			continue
		}
		v := prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: app,
			Name:      name,
			Buckets:   opt.Buckets,
		}, opt.Labels)
		object.histogramVec[name] = v
		object.prometheus.MustRegister(v)
	}
}
