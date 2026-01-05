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

	"go.osspkg.com/console"
	"go.osspkg.com/logx"
)

const formatSyslog = "syslog"

type obj struct {
	file    io.WriteCloser
	handler logx.Logger
	conf    Config
}

func New(tag string, conf Config, handler logx.Logger) io.Closer {
	var err error

	if conf.Format != formatSyslog && len(conf.FilePath) == 0 {
		conf.FilePath = "/dev/stdout"
		conf.Level = logx.LevelError
	}

	log := &obj{
		conf:    conf,
		handler: handler,
	}

	switch conf.Format {
	case formatSyslog:
		network, addr := "", ""
		if uri, err0 := url.Parse(conf.FilePath); err0 == nil {
			network, addr = uri.Scheme, uri.Host
		}
		log.file, err = syslog.Dial(network, addr, syslog.LOG_INFO, tag)
	default:
		log.file, err = os.OpenFile(conf.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	}

	console.FatalIfErr(err, "open log file: %s %s", conf.Format, conf.FilePath)

	log.handler.SetOutput(log.file)
	log.handler.SetLevel(log.conf.Level)

	switch log.conf.Format {
	case "string", "syslog":
		strFmt := logx.NewFormatString()
		strFmt.SetDelimiter(' ')
		log.handler.SetFormatter(strFmt)
	case "json":
		log.handler.SetFormatter(logx.NewFormatJSON())
	}

	return log
}

func (v *obj) Close() error {
	return v.file.Close()
}
