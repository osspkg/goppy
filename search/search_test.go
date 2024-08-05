/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package search_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"go.osspkg.com/casecheck"
	"go.osspkg.com/goppy/v2/search"
)

type TestData struct {
	Line      string
	Data      string
	CreatedAt time.Time
}

type TestDataSearch struct {
	Score     float64
	Line      string
	Data      string
	CreatedAt time.Time
}

func TestUnit_NewSearch(t *testing.T) {
	conf := search.ConfigItem{
		Folder: "/tmp/TestUnit_NewSearch",
		Indexes: []search.ConfigIndex{
			{Name: "demo", Fields: []search.ConfigIndexField{
				{Name: "Line", Type: "text"},
				{Name: "Data", Type: "text"},
				{Name: "CreatedAt", Type: "date"},
			}},
		},
	}

	srv := search.NewSearch(conf)

	casecheck.NoError(t, srv.Generate())
	casecheck.NoError(t, srv.Open())
	defer func() {
		casecheck.NoError(t, os.RemoveAll(conf.Folder))
	}()
	casecheck.NoError(t, srv.Add("demo", "001", TestData{
		Line:      "1",
		Data:      "Hello world",
		CreatedAt: time.Now(),
	}))
	casecheck.NoError(t, srv.Add("demo", "002", TestData{
		Line:      "2",
		Data:      "Happy world",
		CreatedAt: time.Now(),
	}))
	result := make([]TestDataSearch, 0, 10)
	// casecheck.NoError(t, srv.Search("demo", "h* worl*", &result))
	casecheck.NoError(t, srv.Search(context.TODO(), "demo", "Hello world", true, &result))
	fmt.Println(result)
	casecheck.True(t, len(result) == 2)
	casecheck.NoError(t, srv.Delete("demo", "002"))
	result = make([]TestDataSearch, 0, 10)
	casecheck.NoError(t, srv.Search(context.TODO(), "demo", "h* worl*", true, &result))
	casecheck.True(t, len(result) == 1)
	casecheck.NoError(t, srv.Close())
}
