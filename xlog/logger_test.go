/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
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

func TestUnit_NewJSON(t *testing.T) {
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

func TestUnit_NewString(t *testing.T) {
	l := xlog.New()

	xtest.NotNil(t, l)
	l.SetFormatter(xlog.NewFormatString())

	filename, err := os.CreateTemp(os.TempDir(), "test_new_default-*.log")
	xtest.NoError(t, err)

	l.SetOutput(filename)
	l.SetLevel(xlog.LevelDebug)
	xtest.Equal(t, xlog.LevelDebug, l.GetLevel())

	go l.Infof("async %d", 1)
	go l.Warnf("async %d", 2)
	go l.Errorf("async %d", 3)
	go l.Debugf("async %d", 4)

	l.Infof("sync %d", 1)
	l.Warnf("sync %d", 2)
	l.Errorf("sync %d", 3)
	l.Debugf("sync %d", 4)

	l.WithFields(xlog.Fields{"ip": "0.0.0.0"}).Infof("context1")
	l.WithFields(xlog.Fields{"nil": nil}).Infof("context2")
	l.WithFields(xlog.Fields{"func": func() {}}).Infof("context3")

	l.WithField("ip", "0.0.0.0").Infof("context4")
	l.WithField("nil", nil).Infof("context5")
	l.WithField("func", func() {}).Infof("context6")

	l.WithError("err", nil).Infof("context7")
	l.WithError("err", fmt.Errorf("er1")).Infof("context8")

	<-time.After(time.Second * 1)
	l.Close()

	xtest.NoError(t, filename.Close())
	data, err := os.ReadFile(filename.Name())
	xtest.NoError(t, err)
	xtest.NoError(t, os.Remove(filename.Name()))

	sdata := string(data)
	xtest.Contains(t, sdata, "lvl: INF\tmsg: async 1")
	xtest.Contains(t, sdata, "lvl: WRN\tmsg: async 2")
	xtest.Contains(t, sdata, "lvl: ERR\tmsg: async 3")
	xtest.Contains(t, sdata, "lvl: DBG\tmsg: async 4")
	xtest.Contains(t, sdata, "lvl: INF\tmsg: sync 1")
	xtest.Contains(t, sdata, "lvl: WRN\tmsg: sync 2")
	xtest.Contains(t, sdata, "lvl: ERR\tmsg: sync 3")
	xtest.Contains(t, sdata, "lvl: DBG\tmsg: sync 4")
	xtest.Contains(t, sdata, "msg: context1\tctx: [[ip: 0.0.0.0]]")
	xtest.Contains(t, sdata, "msg: context2\tctx: [[nil: <nil>]]")
	xtest.Contains(t, sdata, "msg: context3\tctx: [[func: unsupported field value: (func())")
	xtest.Contains(t, sdata, "msg: context4\tctx: [[ip: 0.0.0.0]]")
	xtest.Contains(t, sdata, "msg: context5\tctx: [[nil: <nil>]]")
	xtest.Contains(t, sdata, "msg: context6\tctx: [[func: unsupported field value: (func())")
	xtest.Contains(t, sdata, "msg: context7\tctx: [[err: <nil>]]")
	xtest.Contains(t, sdata, "msg: context8\tctx: [[err: er1]]")
}

func BenchmarkNewJSON(b *testing.B) {
	b.ReportAllocs()

	ll := xlog.New()
	ll.SetOutput(io.Discard)
	ll.SetLevel(xlog.LevelDebug)
	ll.SetFormatter(xlog.NewFormatJSON())
	wg := iosync.NewGroup()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wg.Background(func() {
				ll.WithFields(xlog.Fields{"a": "b"}).Infof("hello")
				ll.WithField("a", "b").Infof("hello")
				ll.WithError("a", fmt.Errorf("b")).Infof("hello")
			})
		}
	})
	wg.Wait()
	ll.Close()
}

func BenchmarkNewString(b *testing.B) {
	b.ReportAllocs()

	ll := xlog.New()
	ll.SetOutput(io.Discard)
	ll.SetLevel(xlog.LevelDebug)
	ll.SetFormatter(xlog.NewFormatString())
	wg := iosync.NewGroup()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wg.Background(func() {
				ll.WithFields(xlog.Fields{"a": "b"}).Infof("hello")
				ll.WithField("a", "b").Infof("hello")
				ll.WithError("a", fmt.Errorf("b")).Infof("hello")
			})
		}
	})
	wg.Wait()
	ll.Close()
}
