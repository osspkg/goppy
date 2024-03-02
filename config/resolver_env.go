/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package config

import "os"

type envResolver struct{}

func EnvResolver() Resolver {
	return &envResolver{}
}

func (e envResolver) Name() string {
	return "env"
}

func (e envResolver) Value(name string) (string, bool) {
	return os.LookupEnv(name)
}
