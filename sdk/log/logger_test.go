/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package log_test

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.osspkg.com/goppy/sdk/iosync"
	"go.osspkg.com/goppy/sdk/log"
)

func TestUnit_New(t *testing.T) {
	require.NotNil(t, log.Default())

	filename, err := os.CreateTemp(os.TempDir(), "test_new_default-*.log")
	require.NoError(t, err)

	log.SetOutput(filename)
	log.SetLevel(log.LevelDebug)
	require.Equal(t, log.LevelDebug, log.GetLevel())

	go log.Infof("async %d", 1)
	go log.Warnf("async %d", 2)
	go log.Errorf("async %d", 3)
	go log.Debugf("async %d", 4)

	log.Infof("sync %d", 1)
	log.Warnf("sync %d", 2)
	log.Errorf("sync %d", 3)
	log.Debugf("sync %d", 4)

	log.WithFields(log.Fields{"ip": "0.0.0.0"}).Infof("context1")
	log.WithFields(log.Fields{"nil": nil}).Infof("context2")
	log.WithFields(log.Fields{"func": func() {}}).Infof("context3")

	log.WithField("ip", "0.0.0.0").Infof("context4")
	log.WithField("nil", nil).Infof("context5")
	log.WithField("func", func() {}).Infof("context6")

	log.WithError("err", nil).Infof("context7")
	log.WithError("err", fmt.Errorf("er1")).Infof("context8")

	<-time.After(time.Second * 1)
	log.Close()

	require.NoError(t, filename.Close())
	data, err := os.ReadFile(filename.Name())
	require.NoError(t, err)
	require.NoError(t, os.Remove(filename.Name()))

	sdata := string(data)
	require.Contains(t, sdata, `"lvl":"INF","msg":"async 1"`)
	require.Contains(t, sdata, `"lvl":"WRN","msg":"async 2"`)
	require.Contains(t, sdata, `"lvl":"ERR","msg":"async 3"`)
	require.Contains(t, sdata, `"lvl":"DBG","msg":"async 4"`)
	require.Contains(t, sdata, `"lvl":"INF","msg":"sync 1"`)
	require.Contains(t, sdata, `"lvl":"WRN","msg":"sync 2"`)
	require.Contains(t, sdata, `"lvl":"ERR","msg":"sync 3"`)
	require.Contains(t, sdata, `"msg":"context1","ctx":{"ip":"0.0.0.0"}`)
	require.Contains(t, sdata, `"msg":"context2","ctx":{"nil":null}`)
	require.Contains(t, sdata, `"msg":"context3","ctx":{"func":"unsupported field value: (func())`)
	require.Contains(t, sdata, `"msg":"context4","ctx":{"ip":"0.0.0.0"}`)
	require.Contains(t, sdata, `"msg":"context5","ctx":{"nil":null}`)
	require.Contains(t, sdata, `"msg":"context6","ctx":{"func":"unsupported field value: (func())`)
	require.Contains(t, sdata, `"msg":"context7","ctx":{"err":null}`)
	require.Contains(t, sdata, `"msg":"context8","ctx":{"err":"er1"}`)
}

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()

	ll := log.New()
	ll.SetOutput(io.Discard)
	ll.SetLevel(log.LevelDebug)
	wg := iosync.NewGroup()

	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		wg.Background(func() {
			for p.Next() {
				ll.WithFields(log.Fields{"a": "b"}).Infof("hello")
				ll.WithField("a", "b").Infof("hello")
				ll.WithError("a", fmt.Errorf("b")).Infof("hello")
			}
		})
	})
	wg.Wait()
	ll.Close()
}
