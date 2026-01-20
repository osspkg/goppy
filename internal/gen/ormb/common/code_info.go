/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package common

type CodeInfo struct {
	FilePath  string
	PkgName   string
	ModelName string
	Imports   []Import
}

type Import struct {
	Name string
	Pkg  string
}
