/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package metrics

import (
	"fmt"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"go.osspkg.com/goppy/syscall"
)

func fatal(msg string, args ...interface{}) {
	t := syscall.Trace(1000)
	fmt.Fprintf(os.Stderr, msg+t, args...)
	os.Exit(1)
}

func buildPrometheusLabels(keysVals []string) prometheus.Labels {
	if len(keysVals)%2 != 0 {
		fatal("Error parsing names and values for labels, an odd number is specified: %+v", keysVals)
	}
	result := prometheus.Labels{}
	for i := 0; i < len(keysVals); i += 2 {
		result[keysVals[i]] = keysVals[i+1]
	}
	return result
}
