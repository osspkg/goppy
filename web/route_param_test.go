/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"testing"

	"go.osspkg.com/goppy/xtest"
)

func TestHasMatcher(t *testing.T) {
	tests := []struct {
		name string
		args string
		want bool
	}{
		{name: "Case1", args: `test-{id:\d+}`, want: true},
		{name: "Case2", args: `test-id:\d+`, want: false},
		{name: "Case3", args: `test-{id}`, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasParamMatch(tt.args); got != tt.want {
				t.Errorf("HasMatcher() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnit_NewMatcher(t *testing.T) {
	mt := newParamMatch()

	tests1 := []struct {
		name    string
		args    string
		wantErr bool
	}{
		{name: "c1", args: `page-{id}-{title:[\]}`, wantErr: true},
		{name: "c2", args: `page-{id:\d+}-{title2:[0-9]+}`, wantErr: false},
		{name: "c3", args: `page-{id:\d+}-{title1:[a-zA-Z]+}`, wantErr: false},
		{name: "c4", args: `page-{id:\d+}-{title3:.+}`, wantErr: false},
		{name: "c5", args: `page-{id:\d+}-{title5:.*}`, wantErr: false},
	}
	for _, tt := range tests1 {
		t.Run(tt.name, func(t *testing.T) {
			err := mt.Add(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Matcher.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	type args struct {
		vv string
		vr uriParamData
	}
	tests2 := []struct {
		name  string
		args  args
		want  string
		want1 bool
	}{
		{
			name: "c6",
			args: args{
				vv: "hello",
				vr: uriParamData{},
			},
			want:  "",
			want1: false,
		},
		{
			name: "c7",
			args: args{
				vv: "page--",
				vr: uriParamData{},
			},
			want:  "",
			want1: false,
		},
		{
			name: "c8",
			args: args{
				vv: "page-123-Hello",
				vr: uriParamData{"id": "123", "title1": "Hello"},
			},
			want:  `page-{id:\d+}-{title1:[a-zA-Z]+}`,
			want1: true,
		},
		{
			name: "c9",
			args: args{
				vv: "page-123-0000",
				vr: uriParamData{"id": "123", "title2": "0000"},
			},
			want:  `page-{id:\d+}-{title2:[0-9]+}`,
			want1: true,
		},
		{
			name: "c10",
			args: args{
				vv: "page-123-bb-88",
				vr: uriParamData{"id": "123", "title3": "bb-88"},
			},
			want:  `page-{id:\d+}-{title3:.+}`,
			want1: true,
		},
		{
			name: "c11",
			args: args{
				vv: "page-123-",
				vr: uriParamData{"id": "123"},
			},
			want:  `page-{id:\d+}-{title5:.*}`,
			want1: true,
		},
	}
	for _, tt := range tests2 {
		t.Run(tt.name, func(t *testing.T) {
			params := uriParamData{}
			got, got1 := mt.Match(tt.args.vv, params)
			if got != tt.want {
				t.Errorf("Matcher.Match() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Matcher.Match() got1 = %v, want %v", got1, tt.want1)
			}
			xtest.Equal(t, tt.args.vr, params, "Matcher.Match() params = %v, want %v", params, tt.args.vr)
		})
	}
}

func TestUnit_NewMatcher1(t *testing.T) {
	mt := newParamMatch()
	xtest.NoError(t, mt.Add(`{id}`))
	params := uriParamData{}
	path, ok := mt.Match("bbb", params)
	xtest.True(t, ok)
	xtest.Equal(t, `{id}`, path)
	xtest.Equal(t, uriParamData{"id": "bbb"}, params)
}
