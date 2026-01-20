/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"go.osspkg.com/logx"
)

func buildPrometheusLabels(keysVals []string) (prometheus.Labels, *fatalMessage) {
	if len(keysVals)%2 != 0 {
		return nil, fatal("Error parsing names and values for labels, an odd number is specified: %+v", keysVals)
	}
	result := prometheus.Labels{}
	for i := 0; i < len(keysVals); i += 2 {
		result[keysVals[i]] = keysVals[i+1]
	}
	return result, nil
}

const fatalMsg = "Fail Metric"

type fatalMessage struct {
	msg string
}

func fatal(msg string, args ...any) *fatalMessage {
	return &fatalMessage{msg: fmt.Sprintf(msg, args...)}
}

func (v *fatalMessage) Error() string     { return v.msg }
func (v *fatalMessage) Inc()              { logx.Error(fatalMsg, "err", v.msg) }
func (v *fatalMessage) Dec()              { logx.Error(fatalMsg, "err", v.msg) }
func (v *fatalMessage) SetToCurrentTime() { logx.Error(fatalMsg, "err", v.msg) }
func (v *fatalMessage) Add(float64)       { logx.Error(fatalMsg, "err", v.msg) }
func (v *fatalMessage) Set(float64)       { logx.Error(fatalMsg, "err", v.msg) }
func (v *fatalMessage) Sub(float64)       { logx.Error(fatalMsg, "err", v.msg) }
func (v *fatalMessage) Observe(float64)   { logx.Error(fatalMsg, "err", v.msg) }
