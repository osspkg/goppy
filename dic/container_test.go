/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dic_test

import (
	"errors"
	"testing"

	"go.osspkg.com/xc"

	"go.osspkg.com/goppy/v3/dic"
)

// --- Вспомогательные типы для тестов ---

type MockConnector struct {
	startErr error
	stopErr  error
	started  bool
	stopped  bool
}

func (m *MockConnector) Name() string                 { return "MockConnector" }
func (m *MockConnector) Priority() int                { return 0 }
func (m *MockConnector) Apply(arg any)                {}
func (m *MockConnector) OnStart(ctx xc.Context) error { m.started = true; return m.startErr }
func (m *MockConnector) OnStop() error                { m.stopped = true; return m.stopErr }

type TestInterface interface{ Do() string }
type TestImpl struct{ Name string }

func (t *TestImpl) Do() string { return t.Name }

type TestImpl2 struct{ Name string }

func (t *TestImpl2) Do() string { return t.Name }

type StructWithDeps struct {
	Service TestInterface // Экспортируемое поле для инъекции
	secret  string        // Неэкспортируемое поле (должно быть пропущено)
}

type CyclicA struct{ B *CyclicB }
type CyclicB struct{ A *CyclicA }

// --- Сами тесты ---

func TestUnit_Container_Lifecycle(t *testing.T) {
	//dic.ShowDebug(true)
	ctx := xc.New()

	t.Run("Full Success Flow", func(t *testing.T) {
		c := dic.New()
		conn := &MockConnector{}
		_ = c.BrokerRegister(conn)

		// Регистрация константы, функции и структуры
		_ = c.Register("some_string")
		_ = c.Register(func(s string) TestInterface {
			return &TestImpl{Name: s}
		})
		_ = c.Register(StructWithDeps{})

		if err := c.Register("some_string"); err == nil {
			t.Error("register duplicate type should be failed")
		}

		if err := c.Start(ctx); err != nil {
			t.Fatalf("failed to start: %v", err)
		}

		if !conn.started {
			t.Error("connector should be started")
		}

		// Проверка повторного старта (ошибка)
		if err := c.Start(ctx); !errors.Is(err, dic.ErrDepAlreadyRunning) {
			t.Errorf("expected ErrDepAlreadyRunning, got %v", err)
		}

		// Проверка регистрации после старта (ошибка)
		if err := c.Register(1); !errors.Is(err, dic.ErrDepAlreadyRunning) {
			t.Error("expected error when registering after start")
		}

		_ = c.Stop()
		if !conn.stopped {
			t.Error("connector should be stopped")
		}
	})

	t.Run("Connector Start Error", func(t *testing.T) {
		c := dic.New()
		errStart := errors.New("start failed")
		_ = c.BrokerRegister(&MockConnector{startErr: errStart})
		if err := c.Start(ctx); !errors.Is(err, errStart) {
			t.Errorf("expected %v, got %v", errStart, err)
		}
	})
}

func TestUnit_Container_BreakPointAndInvoke(t *testing.T) {
	//dic.ShowDebug(true)
	ctx := xc.New()

	t.Run("BreakPoint Validation", func(t *testing.T) {
		c := dic.New()
		_ = c.Register(func() TestInterface { return &TestImpl{Name: "live"} })

		if err := c.BreakPoint(123); !errors.Is(err, dic.ErrBreakPointType) {
			t.Error("breakpoint should only accept functions")
		}
		// Валидный breakpoint
		if err := c.BreakPoint(func(ti TestInterface) {}); err != nil {
			t.Errorf("failed to set breakpoint: %v", err)
		}
	})

	t.Run("Invoke before start", func(t *testing.T) {
		c := dic.New()
		_ = c.Register(func() TestInterface { return &TestImpl{Name: "live"} })

		err := c.Invoke(func(ti TestInterface) {})
		if !errors.Is(err, dic.ErrDepNotRunning) {
			t.Error("invoke should fail if container is not running")
		}
	})

	t.Run("Invoke valid/invalid", func(t *testing.T) {
		c := dic.New()
		_ = c.Register(func() TestInterface { return &TestImpl{Name: "live"} })
		_ = c.Start(ctx)

		// Не функция
		if err := c.Invoke("not a func"); !errors.Is(err, dic.ErrInvokeType) {
			t.Error("invoke should only accept functions")
		}

		// Успешный вызов
		called := false
		err := c.Invoke(func(ti TestInterface) {
			if ti.Do() == "live" {
				called = true
			}
		})
		if err != nil || !called {
			t.Errorf("invoke failed: %v", err)
		}

		// Функция возвращает ошибку
		customErr := errors.New("fail")
		err = c.Invoke(func() error { return customErr })
		if !errors.Is(err, customErr) {
			t.Errorf("expected %v, got %v", customErr, err)
		}
	})
}

func TestUnit_Container_InitializationLogic(t *testing.T) {
	ctx := xc.New()

	t.Run("Function multi-return and error", func(t *testing.T) {
		c := dic.New()
		customErr := errors.New("factory failed")

		// Регистрация функции возвращающей ошибку
		_ = c.Register(func() (int, error) {
			return 0, customErr
		})

		err := c.Start(ctx)
		if !errors.Is(err, customErr) {
			t.Errorf("expected factory error, got %v", err)
		}
	})

	t.Run("Missing dependencies", func(t *testing.T) {
		c := dic.New()
		// Зависит от string, который не зарегистрирован
		_ = c.Register(func(s string) int { return 1 })

		err := c.Start(ctx)
		if err == nil {
			t.Error("expected error due to missing dependency")
		}
	})

	t.Run("Cyclic Dependency", func(t *testing.T) {
		c := dic.New()
		// Регистрируем структуру саму на себя или цикл (в твоем случае Kahn Graph должен найти цикл)
		_ = c.Register(func(a *CyclicA) *CyclicB { return &CyclicB{A: a} })
		_ = c.Register(func(b *CyclicB) *CyclicA { return &CyclicA{B: b} })

		err := c.Start(ctx)
		if err == nil {
			t.Error("expected error due to cycle in graph")
		}
	})
}

func TestUnit_Container_Collections(t *testing.T) {
	//dic.ShowDebug(true)
	ctx := xc.New()

	t.Run("Interface Slice Injection", func(t *testing.T) {
		c := dic.New()
		_ = c.Register(func() *TestImpl { return &TestImpl{Name: "a"} })
		_ = c.Register(func() *TestImpl2 { return &TestImpl2{Name: "b"} })

		var result []TestInterface
		_ = c.Register(func(list []TestInterface) bool {
			result = list
			return true
		})

		_ = c.Start(ctx)
		if len(result) != 2 {
			t.Errorf("expected 2 items in collection, got %d", len(result))
		}
	})
}
