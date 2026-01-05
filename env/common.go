/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package env

import "os"

type (
	AppName        string
	AppVersion     string
	AppDescription string

	AppInfo struct {
		AppName        AppName
		AppVersion     AppVersion
		AppDescription AppDescription
	}
)

func NewAppInfo() AppInfo {
	return AppInfo{
		AppName:        "",
		AppVersion:     "",
		AppDescription: "",
	}
}

func Get(key, defaultValue string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	return v
}
