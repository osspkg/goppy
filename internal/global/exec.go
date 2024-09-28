/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package global

import (
	"bytes"
	"context"
	"io"
	"os"

	"go.osspkg.com/console"
	"go.osspkg.com/events"
	"go.osspkg.com/ioutils/fs"
	"go.osspkg.com/ioutils/shell"
)

type logger struct {
	Out     io.Writer
	Replace map[string]string
}

func (v *logger) Write(b []byte) (int, error) {
	n := len(b)
	if v.Replace != nil {
		for key, value := range v.Replace {
			b = bytes.ReplaceAll(b, []byte(key), []byte(value))
		}
	}
	_, err := v.Out.Write(b)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func Exec(command string) ([]byte, error) {
	ctx, cncl := context.WithCancel(context.Background())
	go events.OnStopSignal(func() {
		cncl()
	})
	sh := shell.New()
	sh.SetDir(fs.CurrentDir())
	err := sh.SetShell("sh", "e", "c")
	console.FatalIfErr(err, "init shell")
	return sh.Call(ctx, command)
}

func ExecPack(list ...string) {
	ctx, cncl := context.WithCancel(context.Background())
	go events.OnStopSignal(func() {
		cncl()
	})
	sh := shell.New()
	sh.SetDir(fs.CurrentDir())
	err := sh.SetShell("sh", "x", "e", "c")
	console.FatalIfErr(err, "init shell")
	out := &logger{
		Out: os.Stdout,
		Replace: map[string]string{
			fs.CurrentDir(): ".",
		},
	}
	err = sh.CallPackageContext(ctx, out, list...)
	console.FatalIfErr(err, "run command")
}
