/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package rpc

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"go.osspkg.com/do"
	"go.osspkg.com/ioutils/fs"
	"go.osspkg.com/ioutils/shell"
	"go.osspkg.com/logx"
	"go.osspkg.com/syncing"

	"go.osspkg.com/goppy/v3/pkg/xc"

	"go.osspkg.com/goppy/v3/plugins/web/jsonrpc"
)

type unixPlugin struct {
	pidPath    string
	socketPath string

	serv shell.TShell
	cli  *jsonrpc.Client

	conf Config

	ctx    context.Context
	cancel context.CancelFunc
	wg     syncing.Group
}

//nolint:unparam
func newUnixPlugin(c Config) (rpcPlugin, error) {
	ctx, cancel := context.WithCancel(context.Background())
	obj := &unixPlugin{conf: c, wg: syncing.NewGroup(ctx), ctx: ctx, cancel: cancel}

	obj.pidPath = fmt.Sprintf("/tmp/%s.pid", uuid.New().String())
	obj.socketPath = fmt.Sprintf("/tmp/%s.sock", uuid.New().String())

	obj.cli = jsonrpc.New("http://rpc.app.local/",
		jsonrpc.SetUnixSocket(obj.socketPath),
		jsonrpc.SetHeader("X-App-Name", c.Name),
	)

	obj.serv = shell.New()
	obj.serv.UseOSEnv(false)
	obj.serv.SetDir(fs.CurrentDir())
	obj.serv.SetEnv("UNIX_SOCKET_PATH", obj.socketPath)

	return obj, nil
}

func (p *unixPlugin) Start(gx context.Context, opts map[string]string) error {

	cmd := []string{p.conf.Path}
	for k, v := range do.JoinMap(p.conf.Options, opts) {
		cmd = append(cmd, fmt.Sprintf("--%s=%s", k, v))
	}

	p.wg.Background(p.conf.Name, func(bx context.Context) {
		ctx, cncl := xc.Join(bx, gx)
		defer cncl()

		lw := &logWriter{name: p.conf.Name}
		fullCmd := strings.Join(cmd, " ")

		for {
			select {
			case <-ctx.Done():
				return
			default:
				os.Remove(p.socketPath) //nolint:errcheck
				if err := p.serv.CallContext(ctx, lw, fullCmd); err != nil {
					logx.Error(
						"Unix Plugin",
						"err", err,
						"name", p.conf.Name,
					)
				}
			}
		}

	})

	return nil
}

func (p *unixPlugin) Stop() error {
	p.cancel()
	p.wg.Wait()
	return nil
}

func (p *unixPlugin) Call(ctx context.Context, method string, params, result any) error {
	in := jsonrpc.ModelAdapter[any]{Data: params}
	out := jsonrpc.ModelAdapter[any]{Data: result}

	return p.cli.Call(ctx, method, in, &out)
}

type logWriter struct {
	name string
}

func (l *logWriter) Write(p []byte) (n int, err error) {
	logx.Info("Unix Plugin", "name", l.name, "data", string(p))
	return len(p), nil
}
