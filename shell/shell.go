/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package shell

import (
	"context"
	"io"
	"os"
	"os/exec"
	"sync"

	"go.osspkg.com/goppy/errors"
)

type (
	sh struct {
		env   []string
		dir   string
		shell string
		mux   sync.RWMutex
		w     io.Writer
		ch    chan []byte
	}

	Shell interface {
		Close()
		SetEnv(key, value string)
		SetDir(dir string)
		SetShell(shell string)
		SetWriter(w io.Writer)
		CallPackageContext(ctx context.Context, commands ...string) error
		CallContext(ctx context.Context, command string) error
		Call(ctx context.Context, command string) ([]byte, error)
	}
)

func New() Shell {
	v := &sh{
		env:   make([]string, 0),
		dir:   os.TempDir(),
		shell: "/bin/sh",
		w:     &NullWriter{},
		ch:    make(chan []byte, 128),
	}
	go v.Pipe()
	return v
}

func (v *sh) SetEnv(key, value string) {
	v.mux.Lock()
	defer v.mux.Unlock()

	v.env = append(v.env, key+"="+value)
}

func (v *sh) SetDir(dir string) {
	v.mux.Lock()
	defer v.mux.Unlock()

	v.dir = dir
}

func (v *sh) SetShell(shell string) {
	v.mux.Lock()
	defer v.mux.Unlock()

	v.shell = shell
}

func (v *sh) SetWriter(w io.Writer) {
	v.mux.Lock()
	defer v.mux.Unlock()

	v.w = w
}

func (v *sh) Close() {
	v.SetWriter(&NullWriter{})
	close(v.ch)
}

func (v *sh) Pipe() {
	for {
		b, ok := <-v.ch
		if !ok {
			return
		}
		bb := make([]byte, len(b))
		copy(bb, b)
		v.mux.RLock()
		v.w.Write(bb) //nolint:errcheck
		v.mux.RUnlock()
	}
}

func (v *sh) Write(b []byte) (n int, err error) {
	l := len(b)
	select {
	case v.ch <- b:
	default:
	}
	return l, nil
}

func (v *sh) CallPackageContext(ctx context.Context, commands ...string) error {
	for i, command := range commands {
		if err := v.CallContext(ctx, command); err != nil {
			return errors.Wrapf(err, "call command #%d [%s]", i, command)
		}
	}
	return nil
}

func (v *sh) CallContext(ctx context.Context, c string) error {
	v.mux.RLock()
	cmd := exec.CommandContext(ctx, v.shell, "-xec", c, " <&-")
	cmd.Env = append(os.Environ(), v.env...)
	cmd.Dir = v.dir
	cmd.Stdout = v
	cmd.Stderr = v
	v.mux.RUnlock()

	return cmd.Run()
}

func (v *sh) Call(ctx context.Context, c string) ([]byte, error) {
	v.mux.RLock()
	cmd := exec.CommandContext(ctx, v.shell, "-xec", c, " <&-")
	cmd.Env = append(os.Environ(), v.env...)
	cmd.Dir = v.dir
	v.mux.RUnlock()

	return cmd.CombinedOutput()
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type NullWriter struct {
}

func (v *NullWriter) Write(b []byte) (int, error) {
	return len(b), nil
}
