/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package commands

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.osspkg.com/do"
	"go.osspkg.com/events"
	"go.osspkg.com/ioutils/codec"
	"go.osspkg.com/ioutils/fs"

	"go.osspkg.com/goppy/v3/pkg/console"
)

const proxyFilename = ".devproxy.yaml"

type ProxyRule struct {
	Prefix        string            `yaml:"prefix"`
	Forward       string            `yaml:"forward"`
	HeaderSet     map[string]string `yaml:"header_set,omitempty"`
	HeaderDel     []string          `yaml:"header_del,omitempty"`
	HeaderAllowed []string          `yaml:"header_allowed,omitempty"`
}

type ProxyConfig struct {
	Address string      `yaml:"address"`
	Rules   []ProxyRule `yaml:"rules"`
	Default ProxyRule   `yaml:"default"`
}

func (c ProxyConfig) getDefault() proxyRule {
	target, err := url.Parse(c.Default.Forward)
	console.FatalIfErr(err, "Parse `%s`", c.Default.Forward)
	return proxyRule{
		prefix:    c.Default.Prefix,
		target:    target,
		strip:     false,
		setHeader: c.Default.HeaderSet,
		delHeader: c.Default.HeaderDel,
		allowHeader: do.Entries[string, string, struct{}](c.Default.HeaderAllowed, func(s string) (string, struct{}) {
			return strings.ToLower(strings.TrimSpace(s)), struct{}{}
		}),
	}
}

func (c ProxyConfig) getRules() []proxyRule {
	result := make([]proxyRule, 0, len(c.Rules))
	for _, r := range c.Rules {
		target, err := url.Parse(r.Forward)
		console.FatalIfErr(err, "Parse `%s`", r.Forward)
		result = append(result, proxyRule{
			prefix:    r.Prefix,
			target:    target,
			strip:     true,
			setHeader: r.HeaderSet,
			delHeader: r.HeaderDel,
			allowHeader: do.Entries[string, string, struct{}](r.HeaderAllowed, func(s string) (string, struct{}) {
				return strings.ToLower(strings.TrimSpace(s)), struct{}{}
			}),
		})
	}
	return result
}

type proxyRule struct {
	prefix      string
	target      *url.URL
	strip       bool
	setHeader   map[string]string
	delHeader   []string
	allowHeader map[string]struct{}
}

func CmdPROXY() console.CommandGetter {
	return console.NewCommand(func(setter console.CommandSetter) {
		setter.Setup("devproxy", "Run dev proxy server")
		setter.ExecFunc(func() {

			if !fs.FileExist(proxyFilename) {
				config := ProxyConfig{
					Address: "127.0.0.1:8080",
					Rules: []ProxyRule{
						{
							Prefix:  "/api",
							Forward: "http://127.0.0.1:10000",
							HeaderSet: map[string]string{
								"Authorization": "Bearer " + uuid.Nil.String(),
							},
							HeaderDel: []string{"Authorization"},
						},
					},
					Default: ProxyRule{
						Prefix:  "/",
						Forward: "http://0.0.0.0:4200",
						HeaderAllowed: []string{
							"Authorization",
							"Content-Type",
						},
					},
				}
				console.FatalIfErr(codec.FileEncoder(proxyFilename).Encode(config), "create `%s`", proxyFilename)
				console.Fatalf("update `%s`", proxyFilename)
			}

			config := ProxyConfig{}
			console.FatalIfErr(codec.FileEncoder(proxyFilename).Decode(&config), "Read `%s`", proxyFilename)

			defaultRule, rules := config.getDefault(), config.getRules()

			handler := &httputil.ReverseProxy{
				Transport: &http.Transport{
					Proxy:                 http.ProxyFromEnvironment,
					DialContext:           (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
					TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
					ForceAttemptHTTP2:     true,
					MaxIdleConns:          100,
					IdleConnTimeout:       90 * time.Second,
					TLSHandshakeTimeout:   10 * time.Second,
					ExpectContinueTimeout: 1 * time.Second,
				},
				Director: func(req *http.Request) {
					r, ok := proxyMatchRule(req.URL.Path, rules)
					if !ok {
						r = defaultRule
					}

					targetPath := r.target.Path
					reqPath := req.URL.Path
					if r.strip {
						reqPath = strings.TrimPrefix(reqPath, r.prefix)
					}
					req.URL.Path = proxyJoinURLPath(targetPath, reqPath)

					req.URL.Scheme = r.target.Scheme
					req.URL.Host = r.target.Host
					req.Host = r.target.Host

					if r.target.RawQuery != "" {
						if req.URL.RawQuery == "" {
							req.URL.RawQuery = r.target.RawQuery
						} else {
							req.URL.RawQuery = r.target.RawQuery + "&" + req.URL.RawQuery
						}
					}

					if len(r.allowHeader) > 0 {
						for key := range req.Header {
							key = strings.ToLower(key)
							if _, ok := r.allowHeader[key]; !ok {
								req.Header.Del(key)
							}
						}
					}

					if len(r.delHeader) > 0 {
						for _, key := range r.delHeader {
							req.Header.Del(key)
						}
					}

					if len(r.setHeader) > 0 {
						for key, val := range r.setHeader {
							req.Header.Set(key, val)
						}
					}

					console.Infof("forward: [%s] %s -> %s", req.Method, req.URL.Path, req.URL.String())
				},
				ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
					console.WarnIfErr(err, "proxy error")
					http.Error(w, "Bad Gateway", http.StatusBadGateway)
				},
			}

			srv := &http.Server{
				Addr:              config.Address,
				Handler:           handler,
				ReadHeaderTimeout: 1 * time.Second,
			}

			ctx, cancel := context.WithCancel(context.Background())
			go events.OnStopSignal(cancel)

			go func() {
				defer cancel()
				console.Infof("Listening on %s", srv.Addr)
				console.WarnIfErr(srv.ListenAndServe(), "start http server")
			}()

			<-ctx.Done()

			ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			console.WarnIfErr(srv.Shutdown(ctx), "shutdown http server")
		})
	})
}

func proxyMatchRule(path string, rules []proxyRule) (proxyRule, bool) {
	for _, r := range rules {
		if strings.HasPrefix(path, r.prefix) {
			return r, true
		}
	}
	return proxyRule{}, false
}

func proxyJoinURLPath(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
