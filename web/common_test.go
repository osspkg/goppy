/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"reflect"
	"testing"
)

func TestUnit_urlSplit(t *testing.T) {
	type args struct {
		uri string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{name: "Case1", args: args{uri: ""}, want: []string{}},
		{name: "Case2", args: args{uri: "/a/b/"}, want: []string{"a", "b"}},
		{name: "Case3", args: args{uri: "/a/////b/"}, want: []string{"a", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := urlSplit(tt.args.uri); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("split() = %v, want %v", got, tt.want)
			}
		})
	}
}
