/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package custom_type

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSONbInterface interface {
	json.Marshaler
	json.Unmarshaler
}

type JSONb struct {
	Any any
}

func (jb *JSONb) Scan(value any) error {
	if jb.Any == nil || value == nil {
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("type assertion to jsonb failed, got %T", value)
	}

	return json.Unmarshal(b, jb.Any)
}

func (jb *JSONb) Value() (driver.Value, error) {
	if jb.Any == nil {
		return nil, nil
	}

	b, err := json.Marshal(jb.Any)
	if err != nil {
		return nil, err
	}

	return driver.Value(b), nil
}
