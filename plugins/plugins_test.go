/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package plugins_test

import (
	"testing"

	"go.osspkg.com/casecheck"

	"go.osspkg.com/goppy/v3/plugins"
)

func TestKinds_Inject(t *testing.T) {
	k := plugins.Kinds{}
	casecheck.Equal(t, 0, len(k))

	k = k.Inject(0)
	casecheck.Equal(t, 1, len(k))

	k = k.Inject(plugins.Kind{})
	casecheck.Equal(t, 2, len(k))

	k = k.Inject(plugins.Kinds{}.Inject(0))
	casecheck.Equal(t, 3, len(k))

	k = k.Inject(plugins.Inject(1))
	casecheck.Equal(t, 4, len(k))
}
