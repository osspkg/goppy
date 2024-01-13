/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package metrics_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"go.osspkg.com/goppy/env"
	"go.osspkg.com/goppy/metrics"
	"go.osspkg.com/goppy/web"
	"go.osspkg.com/goppy/xc"
	"go.osspkg.com/goppy/xlog"
	"go.osspkg.com/goppy/xnet"
	"go.osspkg.com/goppy/xtest"
)

func TestUnit_NewServer(t *testing.T) {
	app := env.NewAppInfo()
	app.AppName = "app_metrics_test"
	app.AppVersion = "v0.0.0"
	app.AppDescription = "Unit test for metrics server."

	ctx := xc.New()
	logBuff := bytes.NewBuffer(nil)

	log := xlog.New()
	log.SetLevel(xlog.LevelError)
	log.SetOutput(logBuff)
	defer log.Close()

	addr, err := xnet.RandomPort("127.0.0.1")
	xtest.NoError(t, err)
	url := fmt.Sprintf("http://%s/metrics", addr)

	conf := metrics.Config{
		Addr:       addr,
		Counter:    []string{"a1"},
		CounterVec: map[string][]string{"a2": {"l1_1"}},
		Gauge:      []string{"a3"},
		GaugeVec:   map[string][]string{"a4": {"l2_1", "l2_2"}},
		Histogram: map[string][]float64{
			"a5": {.05, .5, 1, 2, 5, 10},
		},
		HistogramVec: map[string]metrics.Buckets{
			"a6": {
				Labels:  []string{"l3_1"},
				Buckets: []float64{.05, .5, 1, 2, 5, 10},
			},
		},
	}

	srv := metrics.New(app, conf, log)
	xtest.NoError(t, err)
	xtest.NoError(t, srv.Up(ctx))

	time.Sleep(2 * time.Second)
	fmt.Println(logBuff.String())

	metrics.Counter("a1").Add(5.912)
	metrics.CounterVec("a2", "l1_1", "v1").Add(0.538)
	metrics.Gauge("a3").Add(10.726)
	metrics.Gauge("a3").Dec()
	metrics.GaugeVec("a4", "l2_1", "v2", "l2_2", "v3").Add(10.579)
	metrics.GaugeVec("a4", "l2_1", "v2", "l2_2", "v4").Dec()
	metrics.Histogram("a5").Observe(2.456)
	metrics.HistogramVec("a6", "l3_1", "v5").Observe(0.123)

	var result bytes.Buffer
	cli := web.NewClientHttp()
	err = cli.Call(ctx.Context(), http.MethodGet, url, nil, &result)
	xtest.NoError(t, err)
	xtest.NoError(t, srv.Down())

	data := result.String()
	xtest.Contains(t, data, "app_metrics_test_a1 5.912")
	xtest.Contains(t, data, "app_metrics_test_a2{l1_1=\"v1\"} 0.538")
	xtest.Contains(t, data, "app_metrics_test_a3 9.726")
	xtest.Contains(t, data, "app_metrics_test_a4{l2_1=\"v2\",l2_2=\"v3\"} 10.579")
	xtest.Contains(t, data, "app_metrics_test_a4{l2_1=\"v2\",l2_2=\"v4\"} -1")
	xtest.Contains(t, data, "app_metrics_test_a5_bucket{le=\"0.05\"} 0")
	xtest.Contains(t, data, "app_metrics_test_a5_bucket{le=\"0.5\"} 0")
	xtest.Contains(t, data, "app_metrics_test_a5_bucket{le=\"1\"} 0")
	xtest.Contains(t, data, "app_metrics_test_a5_bucket{le=\"2\"} 0")
	xtest.Contains(t, data, "app_metrics_test_a5_bucket{le=\"5\"} 1")
	xtest.Contains(t, data, "app_metrics_test_a5_bucket{le=\"10\"} 1")
	xtest.Contains(t, data, "app_metrics_test_a5_bucket{le=\"+Inf\"} 1")
	xtest.Contains(t, data, "app_metrics_test_a5_sum 2.456")
	xtest.Contains(t, data, "app_metrics_test_a5_count 1")
	xtest.Contains(t, data, "app_metrics_test_a6_bucket{l3_1=\"v5\",le=\"0.05\"} 0")
	xtest.Contains(t, data, "app_metrics_test_a6_bucket{l3_1=\"v5\",le=\"0.5\"} 1")
	xtest.Contains(t, data, "app_metrics_test_a6_bucket{l3_1=\"v5\",le=\"1\"} 1")
	xtest.Contains(t, data, "app_metrics_test_a6_bucket{l3_1=\"v5\",le=\"2\"} 1")
	xtest.Contains(t, data, "app_metrics_test_a6_bucket{l3_1=\"v5\",le=\"5\"} 1")
	xtest.Contains(t, data, "app_metrics_test_a6_bucket{l3_1=\"v5\",le=\"10\"} 1")
	xtest.Contains(t, data, "app_metrics_test_a6_bucket{l3_1=\"v5\",le=\"+Inf\"} 1")
	xtest.Contains(t, data, "app_metrics_test_a6_sum{l3_1=\"v5\"} 0.123")
	xtest.Contains(t, data, "app_metrics_test_a6_count{l3_1=\"v5\"} 1")
	xtest.Contains(t, data, "app_metrics_test_build_info{arch=\"amd64\",os=\"linux\",version=\"v0.0.0\"} 1")
}
