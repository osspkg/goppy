/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package xlog

import (
	"bytes"
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
				UnixTime: 123456789,
				Level:    "INF",
				Message:  "Hello",
				Ctx: map[string]interface{}{
					"err": "err\nmsg",
				},
			},
			want:    []byte("lvl=INF\tmsg=\"Hello\"\terr=\"err\\nmsg\"\n"),
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
			if !bytes.Contains(got, tt.want) {
				t.Errorf("Encode() got = %v, want %v", string(got), string(tt.want))
			}
		})
	}
}
