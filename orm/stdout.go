/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"fmt"
	"io"
	"os"
	"time"
)

type stdout struct {
	Writer io.Writer
}

// StdOutWriter simple stdout writer
var StdOutWriter = &stdout{Writer: os.Stdout}

func (s *stdout) currentTime() string {
	return time.Now().Format(time.RFC3339)
}

// Write metric
func (s *stdout) Write(p []byte) (n int, err error) {
	return s.Writer.Write(p)
}

// Metric write metric to log
func (s *stdout) Metric(name string, t time.Duration) {
	fmt.Fprintf(s, "[MTRC] %s - %s: %s\n", s.currentTime(), name, t) //nolint:errcheck
}
