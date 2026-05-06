/*
 *  Copyright (c) 2021-2023 Mikhail Knyazhev <markus621@gmail.com>. All rights reserved.
 *  Use of this source code is governed by a BSD-3-Clause license that can be found in the LICENSE file.
 */

package mime_test

import (
	"testing"

	"go.osspkg.com/goppy/v3/pkg/mime"
)

func TestDetectContentType(t *testing.T) {
	type args struct {
		filename string
		b        []byte
	}
	tests := []struct {
		name  string
		args  args
		want1 string
		want2 string
	}{
		{
			name: "stype.css",
			args: args{
				filename: "/style.css",
				b:        []byte(`import ...`),
			},
			want1: "text/css",
			want2: "application/octet-stream",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mime.DetectByFilename(tt.args.filename); got != tt.want1 {
				t.Errorf("DetectContentType() = %v, want %v", got, tt.want1)
			}
			if got := mime.DetectByContent(tt.args.b); got != tt.want2 {
				t.Errorf("DetectContentType() = %v, want %v", got, tt.want2)
			}
		})
	}
}
