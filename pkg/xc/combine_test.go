/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xc

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestUnit_Join(t *testing.T) {
	c, cancel := Join(context.Background(), context.Background())
	if c == nil {
		t.Fatalf("xc.Join returned nil")
	}

	select {
	case <-c.Done():
		t.Fatalf("<-c.Done() == it should block")
	default:
	}

	if _, ok := c.Deadline(); ok {
		t.Fatalf("c.Deadline() == should not has deadline")
	}

	cancel()
	<-time.After(time.Second)

	select {
	case <-c.Done():
	default:
		t.Fatalf("<-c.Done() it shouldn't block")
	}

	if got, want := fmt.Sprint(c), "xc.Join"; got != want {
		t.Fatalf("xc.Join() = %q want %q", got, want)
	}
}

func TestUnit_JoinTimeout(t *testing.T) {
	ct, _ := context.WithTimeout(context.Background(), 1*time.Second) //nolint: govet
	c, _ := Join(context.Background(), context.Background(), ct)      //nolint: govet
	if c == nil {
		t.Fatalf("xc.Join returned nil")
	}

	select {
	case <-c.Done():
		t.Fatalf("<-c.Done() == it should block")
	default:
	}

	if _, ok := c.Deadline(); !ok {
		t.Fatalf("c.Deadline() == should has deadline")
	}

	<-c.Done()

	select {
	case <-c.Done():
	default:
		t.Fatalf("<-c.Done() it shouldn't block")
	}

	if got, want := fmt.Sprint(c), "xc.Join"; got != want {
		t.Fatalf("xc.Join() = %q want %q", got, want)
	}
}

func TestUnit_JoinValue(t *testing.T) {
	c, cncl := Join(
		SetValue(context.TODO(), "a", 123),
		SetValue(context.TODO(), "b", 321),
	)
	defer cncl()
	if c == nil {
		t.Fatalf("xc.Join returned nil")
	}

	if v, ok := GetValue[int](c, "a"); !ok || v != 123 {
		t.Fatalf("<-c.Value() should have value")
	}
	if v, ok := GetValue[int](c, "b"); !ok || v != 321 {
		t.Fatalf("<-c.Value() should have value")
	}
	if _, ok := GetValue[int](c, "c"); ok {
		t.Fatalf("<-c.Value() should have not value")
	}
	if _, ok := GetValue[string](c, "a"); ok {
		t.Fatalf("<-c.Value() should have not value")
	}
}
