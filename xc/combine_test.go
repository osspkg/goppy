/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xc_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.osspkg.com/goppy/xc"
)

func TestUnit_Join(t *testing.T) {
	c, cancel := xc.Join(context.Background(), context.Background())
	if c == nil {
		t.Fatalf("contexts.Join returned nil")
	}

	select {
	case <-c.Done():
		t.Fatalf("<-c.Done() == it should block")
	default:
	}

	cancel()
	<-time.After(time.Second)

	select {
	case <-c.Done():
	default:
		t.Fatalf("<-c.Done() it shouldn't block")
	}

	if got, want := fmt.Sprint(c), "context.Background.WithCancel"; got != want {
		t.Fatalf("contexts.Join() = %q want %q", got, want)
	}
}
