/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package web

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

var rexUrlParams = regexp.MustCompile(`\{([a-z0-9]+)\:?([^{}]*)\}`)

type paramMatch struct {
	incr    int
	keys    map[string]string
	links   map[string]string
	pattern string
	rex     *regexp.Regexp
}

func newParamMatch() *paramMatch {
	return &paramMatch{
		incr:    1,
		pattern: "",
		keys:    make(map[string]string),
		links:   make(map[string]string),
	}
}

func (v *paramMatch) Add(uri string) error {
	result := "^" + uri + "$"

	for _, pattern := range rexUrlParams.FindAllString(result, -1) {
		key := fmt.Sprintf("k%d", v.incr)
		res := rexUrlParams.FindAllStringSubmatch(pattern, 1)[0]

		rex := ".+"
		if len(res) == 3 && len(res[2]) > 0 {
			rex = res[2]
		}

		result = strings.Replace(result, res[0], "(?P<"+key+">"+rex+")", 1)

		v.links[key] = uri
		v.keys[key] = res[1]
		v.incr++
	}

	if _, err := regexp.Compile(result); err != nil {
		return fmt.Errorf("regex compilation error for `%s`: %w", uri, err)
	}

	pattern := v.pattern
	if len(pattern) != 0 {
		pattern += "|"
	}
	pattern += result

	rex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("regex compilation error for `%s`: %w", uri, err)
	}

	v.rex, v.pattern = rex, pattern

	return nil
}

func (v *paramMatch) Match(uri string, params uriParamData) (string, bool) {
	if v.rex == nil {
		return "", false
	}

	matches := v.rex.FindStringSubmatch(uri)
	if len(matches) == 0 {
		return "", false
	}

	link := ""
	for i, name := range v.rex.SubexpNames() {
		val := matches[i]
		if len(val) == 0 {
			continue
		}

		if l, ok := v.links[name]; ok {
			link = l
		}

		if key, ok := v.keys[name]; ok {
			params[key] = val
		}
	}

	return link, true
}

func hasParamMatch(uri string) bool {
	return rexUrlParams.MatchString(uri)
}

/**********************************************************************************************************************/

type (
	uriParamKey  string
	uriParamData map[string]string
)

func ParamString(r *http.Request, key string) (string, error) {
	if v := r.Context().Value(uriParamKey(key)); v != nil {
		if vv, ok := v.(string); ok {
			return vv, nil
		}
	}
	return "", errFailContextKey
}

func ParamInt(r *http.Request, key string) (int64, error) {
	v, err := ParamString(r, key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(v, 10, 64)
}

func ParamFloat(r *http.Request, key string) (float64, error) {
	v, err := ParamString(r, key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(v, 64)
}
