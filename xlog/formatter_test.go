/*
 *  Copyright (c) 2022-2023 Mikhail Knyazhev <markus621@yandex.ru>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xlog

import (
	"reflect"
	"testing"
)

func TestUnit_FormatString_Encode(t *testing.T) {
	tests := []struct {
		name    string
		args    *Message
		want    []byte
		wantErr bool
	}{
		{
			name: "Case1",
			args: &Message{
				Time:    123456789,
				Level:   "INF",
				Message: "Hello",
				Ctx: map[string]interface{}{
					"err": "err msg",
				},
			},
			want:    []byte("time: 123456789\tlvl: INF\tmsg: Hello\tctx: [[err: err msg]]\n"),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fo := NewFormatString()
			got, err := fo.Encode(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Encode() got = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
