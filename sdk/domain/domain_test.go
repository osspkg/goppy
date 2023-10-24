/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package domain_test

import (
	"fmt"
	"testing"

	"go.osspkg.com/goppy/sdk/domain"
)

func TestUnit_Level(t *testing.T) {
	type args struct {
		s     string
		level int
	}
	tests := []struct {
		args args
		want string
	}{
		{
			args: args{
				s:     "www.domain.ltd",
				level: 1,
			},
			want: "ltd",
		},
		{
			args: args{
				s:     "www.domain.ltd",
				level: 2,
			},
			want: "domain.ltd",
		},
		{
			args: args{
				s:     "www.domain.ltd",
				level: 10,
			},
			want: "www.domain.ltd",
		},
		{
			args: args{
				s:     "www.domain.ltd.",
				level: 1,
			},
			want: "ltd.",
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Case %d", i), func(t *testing.T) {
			if got := domain.Level(tt.args.s, tt.args.level); got != tt.want {
				t.Errorf("DomainLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Benchmark_Level(b *testing.B) {
	d := "www.domain.ltd."
	e := "domain.ltd."

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if got := domain.Level(d, 2); got != e {
			b.Errorf("Level() = %v, want %v", got, e)
		}
	}
}
