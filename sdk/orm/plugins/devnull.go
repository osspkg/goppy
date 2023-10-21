/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package plugins

import (
	"io"

	"github.com/osspkg/goppy/sdk/log"
)

var (
	DevNullLog    log.Logger     = &devNullLogger{}
	DevNullMetric MetricExecutor = new(devNullMetric)
)

type devNullMetric struct{}

func (devNullMetric) ExecutionTime(_ string, call func()) { call() }

type devNullLogger struct{}

func (devNullLogger) SetOutput(io.Writer)                            {}
func (devNullLogger) Fatalf(string, ...interface{})                  {}
func (devNullLogger) Errorf(string, ...interface{})                  {}
func (devNullLogger) Warnf(string, ...interface{})                   {}
func (devNullLogger) Infof(string, ...interface{})                   {}
func (devNullLogger) Debugf(string, ...interface{})                  {}
func (devNullLogger) SetLevel(v uint32)                              {}
func (devNullLogger) Close()                                         {}
func (devNullLogger) GetLevel() uint32                               { return 0 }
func (v devNullLogger) WithFields(_ log.Fields) log.Writer           { return v }
func (v devNullLogger) WithField(_ string, _ interface{}) log.Writer { return v }
func (v devNullLogger) WithError(_ string, _ error) log.Writer       { return v }
