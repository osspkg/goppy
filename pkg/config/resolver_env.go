/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package config

import "os"

type envResolver struct{}

func NewEnvResolver() Resolver {
	return &envResolver{}
}

func (e *envResolver) Name() string {
	return "env"
}

func (e *envResolver) Resolve(keys ...string) (map[string][]byte, error) {
	out := make(map[string][]byte, len(keys))

	for _, key := range keys {
		if val, ok := os.LookupEnv(key); ok {
			out[key] = []byte(val)
		}
	}

	return out, nil
}
