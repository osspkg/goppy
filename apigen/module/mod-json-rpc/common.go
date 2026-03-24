/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package mod_json_rpc

import "strings"

const (
	transportName = "JSONRPC%sTransport"

	modelNameReq = "jsonrpc%sModelRequest"
	modelNameRes = "jsonrpc%sModelResponse"
)

const (
	jsonGenComment = "easyjson:json"
)

func ignoreModelParam(tmpl, pt, pp string) bool {
	if tmpl == modelNameRes {
		switch { //nolint:staticcheck
		case pt == "error":
			return true
		default:
		}
	}

	if tmpl == modelNameReq {
		switch {
		case pp == "context" && pt == "Context":
			return true
		default:
		}
	}

	return false
}

func noBodyParam(vals []string) bool {
	for _, val := range vals {
		i := strings.Index(val, ":")
		if i == -1 {
			continue
		}
		switch strings.ToLower(val[0:i]) {
		case "cookie", "header":
			return true
		default:
		}
	}
	return false
}
