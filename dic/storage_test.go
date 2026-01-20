/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dic

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"go.osspkg.com/xc"
)

const pkgPath = "go.osspkg.com/goppy/v3/dic"

// Подготовка тестовой среды: интерфейсы и реализации
type MockConnector interface {
	Connect() bool
}

type MyService struct{ ID int }

func (s *MyService) Connect() bool { return true }

type AnotherService struct{}

func (s AnotherService) Connect() bool { return true }

func TestUnit_ObjectFromAny(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"nil value", nil, "nil"},
		{"base int", 42, "int"},
		{"struct", MyService{ID: 1}, pkgPath + ".MyService"},
		{"pointer", &MyService{ID: 1}, "*" + pkgPath + ".MyService"},
		{"error interface", errors.New("err"), "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := objectFromAny(tt.input)
			if obj.Address != tt.expected {
				t.Errorf("expected address %q, got %q", tt.expected, obj.Address)
			}

			if tt.input == nil {
				if obj.Type != nil {
					t.Error("expected nil type for nil input")
				}
			} else {
				if obj.Type != reflect.TypeOf(tt.input) {
					t.Errorf("type mismatch: expected %v, got %v", reflect.TypeOf(tt.input), obj.Type)
				}
			}
		})
	}
}

func TestUnit_Storage_SetAndGet(t *testing.T) {
	s := newStorage()
	val := &MyService{ID: 100}
	obj := objectFromAny(val)

	t.Run("successful set", func(t *testing.T) {
		err := s.Set(obj)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("get existing", func(t *testing.T) {
		found, err := s.Get(obj.Address)
		if err != nil {
			t.Fatalf("failed to get object: %v", err)
		}
		if found.Value.Interface() != val {
			t.Error("returned object value differs from original")
		}
	})

	t.Run("prevent duplicate init", func(t *testing.T) {
		// Попытка установить объект с тем же адресом, где Value уже валидно
		err := s.Set(obj)
		if err == nil {
			t.Error("expected error for already initiated dependency, got nil")
		}
	})

	t.Run("allow update uninitiated", func(t *testing.T) {
		// Регистрация типа без значения (как делает container.append)
		addr := "uninitiated.service"
		uninit := &object{Address: addr, Type: reflect.TypeOf(MyService{})}

		_ = s.Set(uninit)

		// Теперь "инициализируем" его
		initObj := &object{
			Address: addr,
			Type:    reflect.TypeOf(MyService{}),
			Value:   reflect.ValueOf(MyService{ID: 1}),
		}

		err := s.Set(initObj)
		if err != nil {
			t.Errorf("should allow setting value for uninitiated object, got: %v", err)
		}
	})

	t.Run("get non-existent", func(t *testing.T) {
		_, err := s.Get("missing.key")
		if err == nil {
			t.Error("expected error for missing key")
		}
	})
}

func TestUnit_Storage_GetCollection(t *testing.T) {
	s := newStorage()

	// 1. Регистрируем одиночные объекты, реализующие MockConnector
	s1 := &MyService{ID: 1}
	s2 := AnotherService{}

	_ = s.Set(objectFromAny(s1))
	_ = s.Set(objectFromAny(s2))

	// 2. Регистрируем слайс интерфейсов (твой storage.go умеет их "распаковывать")
	sliceOfConnectors := []MockConnector{&MyService{ID: 2}}
	_ = s.Set(objectFromAny(sliceOfConnectors))

	t.Run("collect all implementations", func(t *testing.T) {
		// Запрашиваем тип []MockConnector
		target := reflect.TypeOf([]MockConnector{})
		collectionValue := s.GetCollection(target)

		if collectionValue.Kind() != reflect.Slice {
			t.Fatalf("expected slice, got %v", collectionValue.Kind())
		}

		// Должно быть 3 элемента: s1, s2 и один из слайса
		if collectionValue.Len() != 3 {
			t.Errorf("expected 3 elements, got %d", collectionValue.Len())
		}

		// Проверяем, что каждый элемент реализует интерфейс
		for i := 0; i < collectionValue.Len(); i++ {
			item := collectionValue.Index(i).Interface()
			if _, ok := item.(MockConnector); !ok {
				t.Errorf("item at index %d does not implement MockConnector", i)
			}
		}
	})
}

func TestName(t *testing.T) {
	obj := xc.New()

	fmt.Println(reflect.TypeOf(obj).Kind())
}
