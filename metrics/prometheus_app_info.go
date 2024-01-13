/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	cm "github.com/prometheus/client_model/go"
	"google.golang.org/protobuf/proto"
)

type appInfo struct {
	desc       *prometheus.Desc
	labelPairs []*cm.LabelPair
}

func (v appInfo) Desc() *prometheus.Desc {
	return v.desc
}

func (v appInfo) Write(out *cm.Metric) error {
	out.Label = v.labelPairs
	out.Gauge = &cm.Gauge{Value: proto.Float64(1)}
	return nil
}
