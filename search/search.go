/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package search

import (
	"context"

	"github.com/blevesearch/bleve/v2"
	"go.osspkg.com/syncing"
)

type (
	Searcher interface {
		CreateIndex(name string, fields []string) error
		AddDocuments(name, id string, data ...map[string]string) error
		DeleteDocuments(name string, ids ...string) error
		Search(ctx context.Context, query Query) (*Result, error)
		Close() error
	}

	service struct {
		conf Config
		list *syncing.Map[string, bleve.Index]
	}
)

func New(c Config) Searcher {
	return &service{
		conf: c,
		list: syncing.NewMap[string, bleve.Index](2),
	}
}
