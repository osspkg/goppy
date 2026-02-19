/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package types

type Object struct {
	Name    string
	Methods []Method
	Tags    Tags
}

type Method struct {
	Name      string
	Tags      Tags
	InParams  []Param
	OutParams []Param
}

type Param struct {
	Name      string
	Type      string
	Pkg       string
	Omitempty bool
}

type KV struct {
	Key   string
	Value string
}

type Tags map[string][]string
