/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

	"go.osspkg.com/goppy/v2/env"
	"go.osspkg.com/goppy/v2/plugins"
)

type ConfigMetrics struct {
	Config Config `yaml:"metrics"`
}

func (v *ConfigMetrics) Default() {
	v.Config = Config{
		Addr:       "0.0.0.0:12000",
		Counter:    []string{"default_counter"},
		CounterVec: map[string][]string{"default_counter_vec": {"label"}},
		Gauge:      []string{"default_gauge"},
		GaugeVec:   map[string][]string{"default_gauge_vec": {"label1", "label2"}},
		Histogram: map[string][]float64{
			"default_histogram": prometheus.DefBuckets,
		},
		HistogramVec: map[string]Buckets{
			"default_histogram_vec": {
				Labels:  []string{"label1", "label2", "label3"},
				Buckets: prometheus.DefBuckets,
			},
		},
	}
}

func WithServer() plugins.Plugin {
	return plugins.Plugin{
		Config: &ConfigMetrics{},
		Inject: func(app env.AppInfo, c *ConfigMetrics) Metrics {
			return New(app, c.Config)
		},
	}
}

type Metrics interface {
	AddHandler(path string, ctrl func(http.ResponseWriter, *http.Request), methods ...string)
}
