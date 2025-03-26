/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package models

//go:generate easyjson

//easyjson:json
type Users []User

//easyjson:json
type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

//easyjson:json
type IntArray []int
