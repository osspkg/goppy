/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package orm

import (
	"os"
	"testing"

	"go.osspkg.com/casecheck"
)

func TestUnit_newImMemoryFS(t *testing.T) {
	m := newMemFS([]Migration{
		{
			Tags: []string{"aaa"},
			Data: map[string]string{
				"0010_aaa.sql": "SQL1",
				"0001_aaa.sql": "SQL2",
				"0001_bbb.sql": "SQL3",
			},
		},
	})

	casecheck.True(t, m.Next())
	casecheck.Equal(t, []string{"aaa"}, m.Tags())

	names, err := m.FileNames()
	casecheck.NoError(t, err)
	casecheck.Equal(t, []string{"0001_aaa.sql", "0001_bbb.sql", "0010_aaa.sql"}, names)

	b, err := m.FileData("0001_aaa.sql")
	casecheck.NoError(t, err)
	casecheck.Equal(t, "SQL2", b)

	b, err = m.FileData("0001_bbb.sql")
	casecheck.NoError(t, err)
	casecheck.Equal(t, "SQL3", b)

	b, err = m.FileData("0010_aaa.sql")
	casecheck.NoError(t, err)
	casecheck.Equal(t, "SQL1", b)

	b, err = m.FileData("1111_aaa.sql")
	casecheck.Error(t, err)
	casecheck.Equal(t, "", b)

	casecheck.False(t, m.Next())
	casecheck.False(t, m.Next())
}

func TestUnit_newOSFS(t *testing.T) {
	casecheck.NoError(t, os.RemoveAll("/tmp/test-migr"))
	casecheck.NoError(t, os.MkdirAll("/tmp/test-migr", 0755))
	casecheck.NoError(t, os.WriteFile("/tmp/test-migr/001_aaa.sql", []byte("sql1"), 0755))
	casecheck.NoError(t, os.WriteFile("/tmp/test-migr/001_bbb.sql", []byte("sql2"), 0755))
	casecheck.NoError(t, os.WriteFile("/tmp/test-migr/011_bbb.sql", []byte("sql3"), 0755))

	m := newOSFS([]ConfigMigrateItem{
		{
			Tags: "aaa,bbb",
			Dir:  "/tmp/test-migr",
		},
	})

	casecheck.True(t, m.Next())
	casecheck.Equal(t, []string{"aaa", "bbb"}, m.Tags())

	names, err := m.FileNames()
	casecheck.NoError(t, err)
	casecheck.Equal(t, []string{
		"/tmp/test-migr/001_aaa.sql",
		"/tmp/test-migr/001_bbb.sql",
		"/tmp/test-migr/011_bbb.sql",
	}, names)

	b, err := m.FileData("/tmp/test-migr/001_aaa.sql")
	casecheck.NoError(t, err)
	casecheck.Equal(t, "sql1", b)

	b, err = m.FileData("/tmp/test-migr/001_bbb.sql")
	casecheck.NoError(t, err)
	casecheck.Equal(t, "sql2", b)

	b, err = m.FileData("/tmp/test-migr/011_bbb.sql")
	casecheck.NoError(t, err)
	casecheck.Equal(t, "sql3", b)

	b, err = m.FileData("/tmp/test-migr/111_bbb.sql")
	casecheck.Error(t, err)
	casecheck.Equal(t, "", b)

	casecheck.False(t, m.Next())
	casecheck.False(t, m.Next())

}
