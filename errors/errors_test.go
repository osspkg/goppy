/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package errors_test

import (
	e "errors"
	"strings"
	"testing"

	"go.osspkg.com/goppy/errors"
)

func TestUnit_New(t *testing.T) {
	type args struct {
		message string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "Case1", args: args{message: "hello"}, want: "hello", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.args.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err.Error() != tt.want {
				t.Errorf("New() error = %v, want %v", err.Error(), tt.want)
				return
			}
		})
	}
}

func TestUnit_Wrap(t *testing.T) {
	type args struct {
		msg []error
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "Case1",
			args:    args{msg: nil},
			want:    "",
			wantErr: false,
		},
		{
			name:    "Case2",
			args:    args{msg: []error{errors.New("hello"), e.New("world")}},
			want:    "hello: world",
			wantErr: true,
		},
		{
			name:    "Case3",
			args:    args{msg: []error{errors.New("err1"), e.New("err2"), nil, e.New("err3")}},
			want:    "err1: err2: err3",
			wantErr: true,
		},
		{
			name: "Case4",
			args: args{msg: []error{errors.Wrapf(errors.New("err1"), "err1 message"),
				errors.Wrapf(e.New("err2"), "err2 message"),
				errors.Wrapf(e.New("err3"), "err3 message")}},
			want:    "err1 message: err1: err2 message: err2: err3 message: err3",
			wantErr: true,
		},
		{
			name:    "Case5",
			args:    args{msg: []error{nil, nil, nil}},
			want:    "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.Wrap(tt.args.msg...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Wrap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want {
				t.Errorf("Wrap() error = %v, want %v", err.Error(), tt.want)
				return
			}
		})
	}
}

func TestUnit_WrapMessage(t *testing.T) {
	type args struct {
		cause   error
		message string
		args    []interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Case1",
			args: args{
				cause:   nil,
				message: "err context",
				args:    nil,
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "Case2",
			args: args{
				cause:   e.New("err1"),
				message: "err context",
				args:    nil,
			},
			want:    "err context: err1",
			wantErr: true,
		},
		{
			name: "Case3",
			args: args{
				cause:   e.New("err1"),
				message: "bad ip %s",
				args:    []interface{}{"127.0.0.1"},
			},
			want:    "bad ip 127.0.0.1: err1",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.Wrapf(tt.args.cause, tt.args.message, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Wrapf() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want {
				t.Errorf("Wrapf() error = %v, want %v", err.Error(), tt.want)
				return
			}
		})
	}
}

func TestUnit_CauseUnwrap(t *testing.T) {
	type fields struct {
		cause   error
		message string
	}
	tests := []struct {
		name    string
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name: "Case1",
			fields: fields{
				cause:   e.New("err1"),
				message: "context",
			},
			want:    "err1",
			wantErr: true,
		},
		{
			name: "Case2",
			fields: fields{
				cause:   nil,
				message: "context",
			},
			want:    "err1",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := errors.Wrapf(tt.fields.cause, tt.fields.message)
			err := errors.Cause(v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Cause() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want {
				t.Errorf("Cause() error = %v, want %v", err.Error(), tt.want)
				return
			}
			err = errors.Unwrap(v)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unwrap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.want {
				t.Errorf("Unwrap() error = %v, want %v", err.Error(), tt.want)
				return
			}
		})
	}
}

func TestUnit_Is(t *testing.T) {
	err0 := errors.New("test")
	type args struct {
		err    error
		target error
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Case1", args: args{err: err0, target: err0}, want: true},
		{name: "Case2", args: args{err: errors.Wrapf(err0, "ttt"), target: err0}, want: true},
		{name: "Case3", args: args{err: errors.New("hello"), target: err0}, want: false},
		{name: "Case4", args: args{err: nil, target: err0}, want: false},
		{name: "Case5", args: args{err: errors.New("hello"), target: nil}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.Is(tt.args.err, tt.args.target); got != tt.want {
				t.Errorf("Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnit_Trace(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "Case1",
			err:  errors.New("test"),
			want: "[trace] go.osspkg.com/goppy/errors_test.TestUnit_Trace.func1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.Trace(tt.err, "msg"); got != nil && !strings.Contains(got.Error(), tt.want) {
				t.Errorf("Trace() = %v, want %v", got.Error(), tt.want)
			}
		})
	}
}
