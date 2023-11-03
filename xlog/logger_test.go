/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xlog_test

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"go.osspkg.com/goppy/iosync"
	"go.osspkg.com/goppy/xlog"
	"go.osspkg.com/goppy/xtest"
)

func TestUnit_New(t *testing.T) {
	xtest.NotNil(t, xlog.Default())

	filename, err := os.CreateTemp(os.TempDir(), "test_new_default-*.log")
	xtest.NoError(t, err)

	xlog.SetOutput(filename)
	xlog.SetLevel(xlog.LevelDebug)
	xtest.Equal(t, xlog.LevelDebug, xlog.GetLevel())

	go xlog.Infof("async %d", 1)
	go xlog.Warnf("async %d", 2)
	go xlog.Errorf("async %d", 3)
	go xlog.Debugf("async %d", 4)

	xlog.Infof("sync %d", 1)
	xlog.Warnf("sync %d", 2)
	xlog.Errorf("sync %d", 3)
	xlog.Debugf("sync %d", 4)

	xlog.WithFields(xlog.Fields{"ip": "0.0.0.0"}).Infof("context1")
	xlog.WithFields(xlog.Fields{"nil": nil}).Infof("context2")
	xlog.WithFields(xlog.Fields{"func": func() {}}).Infof("context3")

	xlog.WithField("ip", "0.0.0.0").Infof("context4")
	xlog.WithField("nil", nil).Infof("context5")
	xlog.WithField("func", func() {}).Infof("context6")

	xlog.WithError("err", nil).Infof("context7")
	xlog.WithError("err", fmt.Errorf("er1")).Infof("context8")

	<-time.After(time.Second * 1)
	xlog.Close()

	xtest.NoError(t, filename.Close())
	data, err := os.ReadFile(filename.Name())
	xtest.NoError(t, err)
	xtest.NoError(t, os.Remove(filename.Name()))

	sdata := string(data)
	xtest.Contains(t, sdata, `"lvl":"INF","msg":"async 1"`)
	xtest.Contains(t, sdata, `"lvl":"WRN","msg":"async 2"`)
	xtest.Contains(t, sdata, `"lvl":"ERR","msg":"async 3"`)
	xtest.Contains(t, sdata, `"lvl":"DBG","msg":"async 4"`)
	xtest.Contains(t, sdata, `"lvl":"INF","msg":"sync 1"`)
	xtest.Contains(t, sdata, `"lvl":"WRN","msg":"sync 2"`)
	xtest.Contains(t, sdata, `"lvl":"ERR","msg":"sync 3"`)
	xtest.Contains(t, sdata, `"msg":"context1","ctx":{"ip":"0.0.0.0"}`)
	xtest.Contains(t, sdata, `"msg":"context2","ctx":{"nil":null}`)
	xtest.Contains(t, sdata, `"msg":"context3","ctx":{"func":"unsupported field value: (func())`)
	xtest.Contains(t, sdata, `"msg":"context4","ctx":{"ip":"0.0.0.0"}`)
	xtest.Contains(t, sdata, `"msg":"context5","ctx":{"nil":null}`)
	xtest.Contains(t, sdata, `"msg":"context6","ctx":{"func":"unsupported field value: (func())`)
	xtest.Contains(t, sdata, `"msg":"context7","ctx":{"err":null}`)
	xtest.Contains(t, sdata, `"msg":"context8","ctx":{"err":"er1"}`)
}

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()

	ll := xlog.New()
	ll.SetOutput(io.Discard)
	ll.SetLevel(xlog.LevelDebug)
	wg := iosync.NewGroup()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		wg.Background(func() {
			for p.Next() {
				ll.WithFields(xlog.Fields{"a": "b"}).Infof("hello")
				ll.WithField("a", "b").Infof("hello")
				ll.WithError("a", fmt.Errorf("b")).Infof("hello")
			}
		})
	})
	wg.Wait()
	ll.Close()
}
