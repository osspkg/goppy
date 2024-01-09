/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package metrics

type CounterInterface interface {
	// Inc increments the counter by 1. Use Add to increment it by arbitrary
	// non-negative values.
	Inc()
	// Add adds the given value to the counter. It panics if the value is <
	// 0.
	Add(float64)
}

// Counter is a Metric that represents a single numerical value that only ever
// goes up. That implies that it cannot be used to count items whose number can
// also go down, e.g. the number of currently running goroutines. Those
// "counters" are represented by Gauges.
//
// A Counter is typically used to count requests served, tasks completed, errors
// occurred, etc.
func Counter(name string) CounterInterface {
	v, ok := object.counter[name]
	if !ok {
		fatal("Counter with name `%s` not found. Add to config.", name)
	}
	return v
}

// CounterVec is a Collector that bundles a set of Counters that all share the
// same Desc, but have different values for their variable labels. This is used
// if you want to count the same thing partitioned by various dimensions
// (e.g. number of HTTP requests, partitioned by response code and
// method).
func CounterVec(name string, labelNameValue ...string) CounterInterface {
	v, ok := object.counterVec[name]
	if !ok {
		fatal("Counter with name `%s` not found. Add to config.", name)
	}
	return v.With(buildPrometheusLabels(labelNameValue))
}

type GaugeInterface interface {
	// Set sets the Gauge to an arbitrary value.
	Set(float64)
	// Inc increments the Gauge by 1. Use Add to increment it by arbitrary
	// values.
	Inc()
	// Dec decrements the Gauge by 1. Use Sub to decrement it by arbitrary
	// values.
	Dec()
	// Add adds the given value to the Gauge. (The value can be negative,
	// resulting in a decrease of the Gauge.)
	Add(float64)
	// Sub subtracts the given value from the Gauge. (The value can be
	// negative, resulting in an increase of the Gauge.)
	Sub(float64)

	// SetToCurrentTime sets the Gauge to the current Unix time in seconds.
	SetToCurrentTime()
}

// Gauge is a Metric that represents a single numerical value that can
// arbitrarily go up and down.
//
// A Gauge is typically used for measured values like temperatures or current
// memory usage, but also "counts" that can go up and down, like the number of
// running goroutines.
func Gauge(name string) GaugeInterface {
	v, ok := object.gauge[name]
	if !ok {
		fatal("Gauge with name `%s` not found. Add to config.", name)
	}
	return v
}

// GaugeVec is a Collector that bundles a set of Gauges that all share the same
// Desc, but have different values for their variable labels. This is used if
// you want to count the same thing partitioned by various dimensions
// (e.g. number of operations queued, partitioned by user and operation
// type).
func GaugeVec(name string, labelNameValue ...string) GaugeInterface {
	v, ok := object.gaugeVec[name]
	if !ok {
		fatal("GaugeVec with name `%s` not found. Add to config.", name)
	}
	return v.With(buildPrometheusLabels(labelNameValue))
}

type HistogramInterface interface {
	// Observe adds a single observation to the histogram. Observations are
	// usually positive or zero.
	Observe(float64)
}

// A Histogram counts individual observations from an event or sample stream in
// configurable static buckets (or in dynamic sparse buckets as part of the
// experimental Native Histograms, see below for more details). Similar to a
// Summary, it also provides a sum of observations and an observation count.
func Histogram(name string) HistogramInterface {
	v, ok := object.histogram[name]
	if !ok {
		fatal("Histogram with name `%s` not found. Add to config.", name)
	}
	return v
}

// HistogramVec is a Collector that bundles a set of Histograms that all share the
// same Desc, but have different values for their variable labels. This is used
// if you want to count the same thing partitioned by various dimensions
// (e.g. HTTP request latencies, partitioned by status code and method).
func HistogramVec(name string, labelNameValue ...string) HistogramInterface {
	v, ok := object.histogramVec[name]
	if !ok {
		fatal("HistogramVec with name `%s` not found. Add to config.", name)
	}
	return v.With(buildPrometheusLabels(labelNameValue))
}
