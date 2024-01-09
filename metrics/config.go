/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package metrics

type Config struct {
	Addr         string               `yaml:"addr"`
	Counter      []string             `yaml:"counter,omitempty"`
	CounterVec   map[string][]string  `yaml:"counter_vec,omitempty"`
	Gauge        []string             `yaml:"gauge,omitempty"`
	GaugeVec     map[string][]string  `yaml:"gauge_vec,omitempty"`
	Histogram    map[string][]float64 `yaml:"histogram,omitempty"`
	HistogramVec map[string]Buckets   `yaml:"histogram_vec,omitempty"`
}

type Buckets struct {
	Labels  []string  `yaml:"labels"`
	Buckets []float64 `yaml:"buckets"`
}
