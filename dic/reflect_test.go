/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package dic

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

// Тестовые типы для проверки именованных сущностей
type testStruct struct {
	Field int
}
type testInterface interface {
	Method()
}

func TestUnit_ReflectCoverage(t *testing.T) {
	// 1. Тестируем isError
	t.Run("isError", func(t *testing.T) {
		errType := reflect.TypeOf((*error)(nil)).Elem()
		if !isError(errType) {
			t.Error("expected true for error interface")
		}
		if isError(reflect.TypeOf(10)) {
			t.Error("expected false for int")
		}
	})

	// 2. Тестируем isInterfaceCollection
	t.Run("isInterfaceCollection", func(t *testing.T) {
		if !isInterfaceCollection(reflect.TypeOf([]testInterface{})) {
			t.Error("expected true for slice of interfaces")
		}
		if isInterfaceCollection(reflect.TypeOf([]int{})) {
			t.Error("expected false for slice of ints")
		}
		if isInterfaceCollection(reflect.TypeOf(10)) {
			t.Error("expected false for non-slice")
		}
	})

	// 3. Тестируем TypingPointer (все ветки switch и ошибки)
	t.Run("TypingPointer", func(t *testing.T) {
		// Успешный кейс: структура
		s := testStruct{Field: 1}
		res, err := TypingPointer([]any{s}, func(a any) error {
			if reflect.TypeOf(a).Kind() != reflect.Ptr {
				return fmt.Errorf("expected pointer")
			}
			return nil
		})
		if err != nil || len(res) == 0 {
			t.Errorf("failed struct pointer conversion: %v", err)
		}

		// Успешный кейс: указатель
		res, err = TypingPointer([]any{&s}, func(a any) error { return nil })
		if err != nil {
			t.Error(err)
		}

		// Ошибка: неподдерживаемый тип (int)
		_, err = TypingPointer([]any{10}, func(a any) error { return nil })
		if err == nil {
			t.Error("expected error for int type")
		}

		// Ошибка из самого колбэка (ветка Struct)
		_, err = TypingPointer([]any{s}, func(a any) error { return errors.New("call err") })
		if err == nil {
			t.Error("expected error from callback (struct)")
		}

		// Ошибка из самого колбэка (ветка Ptr)
		_, err = TypingPointer([]any{&s}, func(a any) error { return errors.New("call err") })
		if err == nil {
			t.Error("expected error from callback (ptr)")
		}
	})

	// 4. Тестируем ResolveAddress (самый объемный метод)
	t.Run("ResolveAddress", func(t *testing.T) {
		tests := []struct {
			name     string
			t        reflect.Type
			v        reflect.Value
			contains string
		}{
			{"nil", nil, reflect.Value{}, "nil"},
			{"interface error", reflect.TypeOf((*error)(nil)).Elem(), reflect.Value{}, "error"},
			{"interface custom", reflect.TypeOf((*testInterface)(nil)).Elem(), reflect.Value{}, "testInterface"},

			// Функции
			{"func unnamed", reflect.TypeOf(func(int) {}), reflect.Value{}, "func(int)"},
			{"func named with ptr", reflect.TypeOf(func(int) {}), reflect.ValueOf(func(int) {}), "0x"},
			{"func complex sig", reflect.TypeOf(func(a int, b string) (int, error) { return 0, nil }), reflect.Value{}, "func(int, string) (int, error)"},
			{"func multi return", reflect.TypeOf(func() (int, int, int) { return 0, 0, 0 }), reflect.Value{}, "(int, int, int)"},

			// Указатели
			{"ptr error", reflect.TypeOf(new(error)), reflect.Value{}, "error"},
			{"ptr named", reflect.TypeOf(&testStruct{}), reflect.Value{}, "*"},
			{"ptr anonymous", reflect.TypeOf(new(struct{})), reflect.Value{}, "*struct{}"},

			// Коллекции
			{"map", reflect.TypeOf(map[string]int{}), reflect.Value{}, "map[string]int"},
			{"slice", reflect.TypeOf([]int{}), reflect.Value{}, "[]int"},
			{"array", reflect.TypeOf([2]int{}), reflect.Value{}, "[2]int"},

			// Каналы
			{"chan both", reflect.TypeOf(make(chan int)), reflect.Value{}, "chan int"},
			{"chan recv", reflect.TypeOf((<-chan int)(nil)), reflect.Value{}, "<-chan int"},
			{"chan send", reflect.TypeOf((chan<- int)(nil)), reflect.Value{}, "chan<- int"},

			// Структуры
			{"struct empty", reflect.TypeOf(struct{}{}), reflect.Value{}, "struct{}"},
			{"struct with fields", reflect.TypeOf(testStruct{}), reflect.Value{}, "testStruct"},

			// Default
			{"base int", reflect.TypeOf(10), reflect.Value{}, "int"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := ResolveAddress(tt.t, tt.v)
				if !strings.Contains(got, tt.contains) {
					t.Errorf("ResolveAddress() = %q, want to contain %q", got, tt.contains)
				}
			})
		}
	})
}
