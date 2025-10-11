/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package client

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	urlpkg "net/url"
	"time"

	"go.osspkg.com/errors"
	"go.osspkg.com/ioutils/cache"
	"go.osspkg.com/ioutils/data"

	"go.osspkg.com/goppy/v2/auth/signature"
	"go.osspkg.com/goppy/v2/web/client/comparison"
)

type HTTPClient interface {
	Send(ctx context.Context, method, url string, in, out any) error
}

type httpCli struct {
	netDialer    *net.Dialer
	nativeClient *http.Client

	defaultHeaders http.Header
	signStore      cache.Cache[string, signature.Signature]

	types []comparison.Type
}

func NewHTTPClient(opts ...HTTPOption) HTTPClient {
	dial := &net.Dialer{
		Timeout:   15 * time.Second,
		KeepAlive: 60 * time.Second,
	}

	cli := &httpCli{
		netDialer: dial,
		nativeClient: &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					return dial.DialContext(ctx, network, addr)
				},
				MaxIdleConns: 10,
			},
		},
		defaultHeaders: http.Header{},
		signStore:      cache.New[string, signature.Signature](),
	}

	WithProxy("env")(cli)
	WithComparisonType(
		comparison.JSON{},
		comparison.XML{},
		comparison.FORMDATA{},
		comparison.BYTES{},
	)(cli)

	for _, opt := range opts {
		opt(cli)
	}

	return cli
}

func (cli *httpCli) Send(ctx context.Context, method, url string, in, out any) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("http client send panic: %v", e)
		}
	}()

	var host string
	if uri, e := urlpkg.Parse(url); e != nil {
		return fmt.Errorf("http client: failed to parse url: %w", e)
	} else {
		uri.Fragment = ""
		host = uri.Host
		url = uri.String()
	}

	var (
		contentType string
		body                        = data.NewBuffer(128)
		compType    comparison.Type = nil
	)

	if in != nil {
		for _, compType = range cli.types {
			contentType, err = compType.Encode(body, in)

			if err != nil {
				if errors.Is(err, comparison.NoCast) {
					compType = nil
					continue
				}

				return fmt.Errorf("http client: failed to encode request body: %w", err)
			}

			break
		}

		if compType == nil {
			return fmt.Errorf("http client: failed to encode request body: no detect comparison")
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, io.NopCloser(body))
	if err != nil {
		return fmt.Errorf("http client: failed to create request: %w", err)
	}
	req.ContentLength = int64(body.Size())

	req.Header.Set("Connection", "keep-alive")
	for key := range cli.defaultHeaders {
		req.Header.Set(key, cli.defaultHeaders.Get(key))
	}
	if len(contentType) > 0 {
		req.Header.Set("Content-Type", contentType)
	}

	if sign, ok := cli.signStore.Get(host); ok && sign != nil {
		b := make([]byte, 0, body.Size()+len(url))
		b = append(b, url...)

		if req.ContentLength != 0 {
			b = body.Bytes()
		}

		if err = signature.Encode(req.Header, sign, b); err != nil {
			return fmt.Errorf("http client: failed to encode signature: %w", err)
		}
	}

	resp, err := cli.nativeClient.Do(req)
	if err != nil {
		return fmt.Errorf("http client: failed to send request: %w", err)
	}
	defer resp.Body.Close() //nolint:errcheck

	body.Reset()
	if _, err = io.Copy(body, resp.Body); err != nil {
		return fmt.Errorf("http client: failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return &HTTPError{
			Err:         fmt.Errorf("http client: bad status code: %d", resp.StatusCode),
			Code:        resp.StatusCode,
			ContentType: resp.Header.Get("Content-Type"),
			Raw:         body,
		}
	}

	if out != nil {
		for _, compType = range cli.types {
			err = compType.Decode(body, out)
			if errors.Is(err, comparison.NoCast) {
				err = nil
				continue
			}

			break
		}
	}

	if err != nil {
		return &HTTPError{
			Err:         fmt.Errorf("http client: failed to decode response body: %w", err),
			Code:        resp.StatusCode,
			ContentType: resp.Header.Get("Content-Type"),
			Raw:         body,
		}
	}

	return nil
}
