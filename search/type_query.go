/*
 *  Copyright (c) 2022-2026 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package search

import (
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
)

type Query interface {
	getQuery() query.Query
	getFrom() int
	getLimit() int
	getHighlight() bool
	getIndexName() string
	getShowFields() []string
}

type MatchQuery struct {
	Query       string
	IndexName   string
	SearchField string
	ShowFields  []string
	From        int
	Limit       int
	Highlight   bool
}

func (v MatchQuery) getQuery() query.Query {
	q := bleve.NewMatchQuery(v.Query)
	q.SetField(v.SearchField)
	return q
}
func (v MatchQuery) getFrom() int            { return v.From }
func (v MatchQuery) getLimit() int           { return v.Limit }
func (v MatchQuery) getHighlight() bool      { return v.Highlight }
func (v MatchQuery) getIndexName() string    { return v.IndexName }
func (v MatchQuery) getShowFields() []string { return v.ShowFields }

type MatchAllQuery struct {
	IndexName string
	From      int
	Limit     int
}

func (v MatchAllQuery) getQuery() query.Query   { return bleve.NewMatchAllQuery() }
func (v MatchAllQuery) getFrom() int            { return v.From }
func (v MatchAllQuery) getLimit() int           { return v.Limit }
func (v MatchAllQuery) getHighlight() bool      { return false }
func (v MatchAllQuery) getIndexName() string    { return v.IndexName }
func (v MatchAllQuery) getShowFields() []string { return nil }

type MatchPhraseQuery struct {
	Query       []string
	IndexName   string
	SearchField string
	ShowFields  []string
	From        int
	Limit       int
	Highlight   bool
}

func (v MatchPhraseQuery) getQuery() query.Query   { return bleve.NewPhraseQuery(v.Query, v.SearchField) }
func (v MatchPhraseQuery) getFrom() int            { return v.From }
func (v MatchPhraseQuery) getLimit() int           { return v.Limit }
func (v MatchPhraseQuery) getHighlight() bool      { return v.Highlight }
func (v MatchPhraseQuery) getIndexName() string    { return v.IndexName }
func (v MatchPhraseQuery) getShowFields() []string { return v.ShowFields }

type MatchUniversalQuery struct {
	Query      string
	IndexName  string
	ShowFields []string
	From       int
	Limit      int
	Highlight  bool
}

func (v MatchUniversalQuery) getQuery() query.Query   { return bleve.NewQueryStringQuery(v.Query) }
func (v MatchUniversalQuery) getFrom() int            { return v.From }
func (v MatchUniversalQuery) getLimit() int           { return v.Limit }
func (v MatchUniversalQuery) getHighlight() bool      { return v.Highlight }
func (v MatchUniversalQuery) getIndexName() string    { return v.IndexName }
func (v MatchUniversalQuery) getShowFields() []string { return v.ShowFields }
