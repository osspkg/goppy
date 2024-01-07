/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package sqlcommon

import (
	"time"
)

type (
	metric struct {
		metrics MetricWriter
	}
	//MetricExecutor interface
	MetricExecutor interface {
		ExecutionTime(name string, call func())
	}
	//MetricWriter interface
	MetricWriter interface {
		Metric(name string, time time.Duration)
	}
)

// StdOutMetric simple stdout metrig writer
var StdOutMetric = NewMetric(StdOutWriter)

// NewMetric init new metric
func NewMetric(m MetricWriter) MetricExecutor {
	return &metric{metrics: m}
}

// ExecutionTime calculating the execution time
func (m *metric) ExecutionTime(name string, call func()) {
	if m.metrics == nil {
		call()
		return
	}

	t := time.Now()
	call()
	m.metrics.Metric(name, time.Since(t))
}
