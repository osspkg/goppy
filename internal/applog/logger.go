/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package applog

import (
	"io"
	"log/syslog"
	"net/url"
	"os"
	"strings"

	"go.osspkg.com/console"
	"go.osspkg.com/logx"
)

const formatSyslog = "syslog"

type obj struct {
	file io.WriteCloser
}

func New(tag string, conf Config) io.Closer {
	var err error

	if len(conf.FilePath) == 0 {
		conf.FilePath = "/dev/stdout"
		conf.Level = logx.LevelError
	}

	log := &obj{}

	switch conf.Format {
	case "string":
		logx.SetDefault(logx.NewSLogStringAdapter())
	default:
		logx.SetDefault(logx.NewSLogJsonAdapter())
	}

	handler := logx.Default()

	switch {
	case strings.HasPrefix(conf.FilePath, formatSyslog):
		network, addr := "", ""

		sysuri := strings.TrimPrefix(conf.FilePath, formatSyslog)
		sysuri = strings.TrimPrefix(sysuri, "=")
		if len(sysuri) > 0 {
			if uri, err0 := url.Parse(sysuri); err0 == nil && uri.Scheme != "" && uri.Host != "" {
				network, addr = uri.Scheme, uri.Host
			}
		}

		log.file, err = syslog.Dial(network, addr, syslog.LOG_INFO|syslog.LOG_LOCAL0, tag)
	default:
		log.file, err = os.OpenFile(conf.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	}

	console.FatalIfErr(err, "open log file: %s %s", conf.Format, conf.FilePath)

	handler.SetOutput(log.file)
	handler.SetLevel(conf.Level)

	return log
}

func (v *obj) Close() error {
	return v.file.Close()
}
