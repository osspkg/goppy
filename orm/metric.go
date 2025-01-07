/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"time"

	"go.osspkg.com/goppy/v2/metrics"
)

type (
	metric struct {
		name string
	}
	metricExecutor interface {
		ExecutionTime(name string, call func())
	}
)

func newMetric(name string) metricExecutor {
	return &metric{
		name: name,
	}
}

// ExecutionTime calculating the execution time
func (m *metric) ExecutionTime(name string, call func()) {
	t := time.Now()
	call()
	metrics.HistogramVec(m.name, "query", name).Observe(time.Since(t).Seconds())
}
