/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package types

import "go.osspkg.com/gogen/types"

type GlobalModule interface {
	Name() string
	Build(w Writer, m GlobalMeta, value []File) error
}

type FaceModule interface {
	Name() string
	Build(w Writer, m FaceMeta, value Face) error
}

type MethodModule interface {
	Name() string
	Build(w Joiner, m MethodMeta, value Method) error
}

type ParamModule interface {
	Name() string
	Build(w Joiner, m ParamMeta, value Param) error
}

type Writer interface {
	WriteFile(fileName string, tok types.Token) error
}

type Joiner interface {
	Join(toks ...types.Token)
}

type ImportSetter interface {
	Set(string, string)
}

type GlobalMeta struct {
	PkgName string
}

type FaceMeta struct {
	PkgName string
	Import  ImportSetter
}

type MethodMeta struct {
	PkgName string
	Import  ImportSetter
}

type ParamType uint8

const (
	ParamIn  ParamType = 0
	ParamOut ParamType = 1
)

type ParamMeta struct {
	Type     ParamType
	CodeName string
	Import   ImportSetter
	Value    string
	Args     Args
}
