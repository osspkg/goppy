/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package types

import (
	"context"
	"fmt"
)

func StdOut(ctx context.Context, arg []Text) error {
	fmt.Println(arg)
	return nil
}
