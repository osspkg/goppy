/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package plugin_test

import (
	"testing"

	"go.osspkg.com/casecheck"

	"go.osspkg.com/goppy/v3/plugin"
)

func TestKinds_Inject(t *testing.T) {
	k := plugin.Kinds{}
	casecheck.Equal(t, 0, len(k))

	k = k.Inject(0)
	casecheck.Equal(t, 1, len(k))

	k = k.Inject(plugin.Kind{})
	casecheck.Equal(t, 2, len(k))

	k = k.Inject(plugin.Kinds{}.Inject(0))
	casecheck.Equal(t, 3, len(k))

	k = k.Inject(plugin.Inject(1))
	casecheck.Equal(t, 4, len(k))
}
