/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"go.osspkg.com/goppy/ioutil"
	"go.osspkg.com/goppy/web/signature"
)

type (
	ClientHttp struct {
		cli *http.Client

		headers   http.Header
		signStore signature.Storage

		enc func(in interface{}) (body []byte, contentType string, err error)
		dec func(code int, contentType string, body []byte, out interface{}) error
	}
)

func NewClientHttp(opt ...ClientHttpOption) *ClientHttp {
	cli := &ClientHttp{
		cli:     http.DefaultClient,
		headers: make(http.Header),
	}
	ClientHttpOptionSetup("env", 5*time.Second, 100)(cli)
	ClientHttpOptionCodecDefault()(cli)
	for _, option := range opt {
		option(cli)
	}
	return cli
}

func (v *ClientHttp) Call(ctx context.Context, method, uri string, in interface{}, out interface{}) error {
	var (
		ioBody      io.Reader
		b           []byte
		contentType string
		err         error
		u           *url.URL
	)

	if u, err = url.Parse(uri); err != nil {
		return err
	}

	if in != nil {
		if b, contentType, err = v.enc(in); err != nil {
			return err
		}
		ioBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, uri, ioBody)
	if err != nil {
		return err
	}

	req.Header.Set("Connection", "keep-alive")
	for k := range v.headers {
		req.Header.Set(k, v.headers.Get(k))
	}
	if len(contentType) > 0 {
		req.Header.Set("Content-Type", contentType)
	}

	if v.signStore != nil {
		if sign := v.signStore.Get(u.Host); sign != nil {
			signature.Encode(req.Header, sign, b)
		}
	}

	resp, err := v.cli.Do(req) //nolint: bodyclose
	if err != nil {
		return err
	}

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return v.dec(resp.StatusCode, resp.Header.Get("Content-Type"), b, out)
}

/**********************************************************************************************************************/

type ClientHttpOption func(c *ClientHttp)

func ClientHttpOptionCodec(
	enc func(in interface{}) (body []byte, contentType string, err error),
	dec func(code int, contentType string, body []byte, out interface{}) error,
) ClientHttpOption {
	return func(c *ClientHttp) {
		c.enc = enc
		c.dec = dec
	}
}

func ClientHttpOptionCodecDefault() ClientHttpOption {
	return ClientHttpOptionCodec(
		func(in interface{}) (body []byte, contentType string, err error) {
			if v, ok := in.(io.Reader); ok {
				body, err = io.ReadAll(v)
				contentType = "text/plain; charset=utf-8"
				return
			}
			switch v := in.(type) {
			case string:
				return []byte(v), "text/plain; charset=utf-8", nil
			case []byte:
				return v, "", nil
			case json.Marshaler:
				body, err = v.MarshalJSON()
				return body, "application/json; charset=utf-8", err
			default:
				return nil, "", fmt.Errorf("unknown request format %T", in)
			}
		},
		func(code int, _ string, body []byte, out interface{}) error {
			switch code {
			case 200:
				if v, ok := out.(io.Writer); ok {
					_, err := v.Write(body)
					return err
				}
				switch v := out.(type) {
				case *string:
					*v += string(body)
					return nil
				case *[]byte:
					*v = append(*v, body...)
					return nil
				case json.Unmarshaler:
					return v.UnmarshalJSON(body)
				default:
					return fmt.Errorf("unknown response format %T", out)
				}

			default:
				return fmt.Errorf("%d %s", code, http.StatusText(code))
			}
		},
	)
}

func ClientHttpOptionSetup(proxy string, ttl time.Duration, countConn int) ClientHttpOption {
	return func(c *ClientHttp) {
		c.cli.Timeout = ttl
		dial := &net.Dialer{
			Timeout:   15 * ttl,
			KeepAlive: 15 * ttl,
		}
		c.cli.Transport = &http.Transport{
			Proxy:               proxySetup(proxy),
			DialContext:         dial.DialContext,
			MaxIdleConns:        countConn,
			MaxIdleConnsPerHost: countConn,
		}
	}
}

func ClientHttpOptionHeaders(keyVal ...string) ClientHttpOption {
	if len(keyVal)%2 != 0 {
		keyVal = append(keyVal, "")
	}
	return func(c *ClientHttp) {
		for i := 0; i < len(keyVal); i += 2 {
			c.headers.Set(keyVal[i], keyVal[i+1])
		}
	}
}

func ClientHttpOptionAuth(s signature.Storage) ClientHttpOption {
	return func(c *ClientHttp) {
		c.signStore = s
	}
}

func proxySetup(proxy string) func(r *http.Request) (*url.URL, error) {
	if len(proxy) == 0 || proxy == "env" {
		return http.ProxyFromEnvironment
	}
	u, err := url.Parse(proxy)
	if err != nil {
		return func(r *http.Request) (*url.URL, error) {
			return nil, err
		}
	}
	return http.ProxyURL(u)
}
