/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package assets

import (
	"net/http"
	"sort"
	"strings"
	"sync"

	"go.osspkg.com/goppy/v3/pkg/mime"
	"go.osspkg.com/goppy/v3/plugins/web"
)

type Reader interface {
	Get(filename string) ([]byte, string)
	List() []string

	ResponseWrite(w http.ResponseWriter, filename string) error
}

// Cache model
type Cache struct {
	files map[string][]byte
	mux   sync.RWMutex
}

// New init cache
func New() *Cache {
	c := &Cache{}
	c.Reset()
	return c
}

// Reset clean cache
func (c *Cache) Reset() {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.files = make(map[string][]byte)
}

// Set setting data to cache
func (c *Cache) Set(filename string, v []byte) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.files[filename] = v
}

// Get getting file by name
func (c *Cache) Get(filename string) ([]byte, string) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	b, ok := c.files[filename]
	if !ok {
		return nil, ""
	}
	return b, mime.DetectByFilename(filename)
}

func (c *Cache) ResponseWrite(w http.ResponseWriter, filename string) error {
	b, ct := c.Get(filename)
	if b == nil {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	w.Header().Set("Content-Type", ct)
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(b)
	return err
}

type WebHandlerConfig struct {
	DefaultIndex  string
	NotFoundIndex string
	ReplacePrefix map[string]string
}

func (c *Cache) WebHandler(whc WebHandlerConfig) func(ctx web.Ctx) {
	list := make([]string, 0, len(whc.ReplacePrefix)*2)
	for from, to := range whc.ReplacePrefix {
		list = append(list, from, to)
	}
	repl := strings.NewReplacer(list...)
	return func(ctx web.Ctx) {
		filepath := ctx.Request().RequestURI

		switch filepath {
		case "", "/":
			filepath = whc.DefaultIndex
		default:
			filepath = repl.Replace(filepath)
		}

		code := http.StatusOK

		b, ct := c.Get(filepath)
		if b == nil {
			code = http.StatusNotFound
			filepath = whc.NotFoundIndex

			b, ct = c.Get(filepath)
		}

		if b == nil {
			ct = mime.TextHTML
		}

		ctx.Raw(code, ct, b)
	}
}

// List getting all files list
func (c *Cache) List() []string {
	c.mux.RLock()
	defer c.mux.RUnlock()

	result := make([]string, 0, len(c.files))
	for name := range c.files {
		result = append(result, name)
	}

	sort.Strings(result)

	return result
}
