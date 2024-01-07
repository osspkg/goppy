/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package shell_test

import (
	"context"
	"testing"

	"go.osspkg.com/goppy/shell"
)

func TestUnit_ShellCall(t *testing.T) {
	sh := shell.New()
	sh.SetDir("/tmp")
	sh.SetEnv("LANG", "en_US.UTF-8")

	out, err := sh.Call(context.TODO(), "ls -la /tmp")
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Log(string(out))
}
