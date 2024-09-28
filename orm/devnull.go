/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"io"

	"go.osspkg.com/logx"
)

var (
	DevNullLog    logx.Logger    = &devNullLogger{}
	DevNullMetric MetricExecutor = new(devNullMetric)
)

type devNullMetric struct{}

func (devNullMetric) ExecutionTime(_ string, call func()) { call() }

type devNullLogger struct{}

func (devNullLogger) SetOutput(out io.Writer)                   {}
func (devNullLogger) SetFormatter(f logx.Formatter)             {}
func (devNullLogger) SetLevel(v uint32)                         {}
func (devNullLogger) Fatal(message string, args ...interface{}) {}
func (devNullLogger) Error(message string, args ...interface{}) {}
func (devNullLogger) Warn(message string, args ...interface{})  {}
func (devNullLogger) Info(message string, args ...interface{})  {}
func (devNullLogger) Debug(message string, args ...interface{}) {}
