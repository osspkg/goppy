/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package types

import (
	"strings"

	"go.osspkg.com/gogen/golang"
	"go.osspkg.com/gogen/types"
)

func TagSplit(arg string) (string, string, Args) {
	val := strings.SplitN(arg, ":", 3)
	args := make(Args)

	switch len(val) {
	case 0:
		return "", "", args
	case 1:
		return strings.TrimSpace(val[0]), "", args
	default:
		if len(val) > 2 {
			for _, item := range strings.Split(val[2], "|") {
				vals := strings.Split(item, "=")
				if len(vals) == 2 {
					args[strings.TrimSpace(vals[0])] = strings.TrimSpace(vals[1])
				}
			}
		}
		return strings.TrimSpace(val[0]), strings.TrimSpace(val[1]), args
	}
}

type Join struct {
	Tok *golang.Tokens
}

func (j *Join) Join(toks ...types.Token) {
	j.Tok = j.Tok.Join(toks...)
}
