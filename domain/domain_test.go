/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package domain_test

import (
	"fmt"
	"testing"

	"go.osspkg.com/goppy/domain"
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
			want: "ltd.",
		},
		{
			args: args{
				s:     "www.domain.ltd",
				level: 2,
			},
			want: "domain.ltd.",
		},
		{
			args: args{
				s:     "www.domain.ltd",
				level: 10,
			},
			want: "www.domain.ltd.",
		},
		{
			args: args{
				s:     "www.domain.ltd.",
				level: 1,
			},
			want: "ltd.",
		},
		{
			args: args{
				s:     "ltd.",
				level: 3,
			},
			want: "ltd.",
		},
		{
			args: args{
				s:     "www.domain.ltd.",
				level: 0,
			},
			want: ".",
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

func BenchmarkDomainLevel(b *testing.B) {
	address := "www.domain.ltd."
	expected := "domain.ltd."

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if got := domain.Level(address, 2); got != expected {
			b.Errorf("DomainLevel() = %v, want %v", got, expected)
		}
	}
}

func TestUnit_Normalize(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		want    string
		wantErr bool
	}{
		{
			name:    "Case1",
			domain:  "1www.a-aa.com",
			want:    "1www.a-aa.com.",
			wantErr: false,
		},
		{
			name:    "Case2",
			domain:  "1_www.aaa.com",
			want:    "",
			wantErr: true,
		},
		{
			name:    "Case3",
			domain:  "com",
			want:    "com.",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.Normalize(tt.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("Normalize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Normalize() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnit_CountLevels(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want int
	}{
		{
			name: "Case1",
			arg:  "",
			want: 0,
		},
		{
			name: "Case2",
			arg:  "aaa.",
			want: 1,
		},
		{
			name: "Case3",
			arg:  "aaa.bbb.",
			want: 2,
		},
		{
			name: "Case4",
			arg:  "aaa.bbb",
			want: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := domain.CountLevels(tt.arg); got != tt.want {
				t.Errorf("CountLevels() = %v, want %v", got, tt.want)
			}
		})
	}
}
