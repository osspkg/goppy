/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package assets

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
)

func (c *Cache) FromBase64TGZ(v string) error {
	b64, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return err
	}
	return c.FromTGZ(bytes.NewBuffer(b64))
}

func (c *Cache) FromTGZ(r io.Reader) error {
	gzf, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	return c.FromTar(gzf)
}

func (c *Cache) FromTar(r io.Reader) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		b, err := io.ReadAll(tr)
		if err != nil {
			return err
		}
		c.Set(hdr.Name, b)
	}
	return nil
}

func (c *Cache) ToTGZ(w io.Writer) error {
	gw := gzip.NewWriter(w)
	err := c.ToTar(gw)
	if err != nil {
		return err
	}
	if err = gw.Close(); err != nil {
		return err
	}
	return nil
}

func (c *Cache) ToTar(w io.Writer) error {
	tw := tar.NewWriter(w)
	defer tw.Close() //nolint:errcheck

	c.mux.RLock()
	defer c.mux.RUnlock()

	for _, name := range c.List() {
		v, ok := c.files[name]
		if !ok {
			return fmt.Errorf("file not found: %s", name)
		}
		hdr := &tar.Header{
			Name: name,
			Mode: int64(os.ModePerm),
			Size: int64(len(v)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := tw.Write(v); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cache) ToBase64TGZ(w io.Writer) error {
	wc := base64.NewEncoder(base64.StdEncoding, w)
	defer wc.Close() //nolint:errcheck
	return c.ToTGZ(wc)
}
