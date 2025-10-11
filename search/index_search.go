/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
)

func (v *service) Search(ctx context.Context, query Query) (*Result, error) {
	index, ok := v.list.Get(query.getIndexName())
	if !ok {
		return nil, fmt.Errorf("no such index '%s'", query.getIndexName())
	}

	request := bleve.NewSearchRequest(query.getQuery())
	request.IncludeLocations = true
	request.Size = max(query.getLimit(), 1)
	request.From = query.getFrom()
	request.Explain = false
	request.Fields = []string{fieldCreatedAt}

	if query.getHighlight() {
		request.Highlight = bleve.NewHighlight()
	} else {
		request.Fields = append(request.Fields, query.getShowFields()...)

		if len(request.Fields) <= 1 {
			if fields, err := index.Fields(); err == nil {
				request.Fields = append(request.Fields, fields...)
			}
		}
	}

	searchResult, err := index.SearchInContext(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	result := &Result{
		Total:     searchResult.Total,
		Took:      searchResult.Took,
		MaxScore:  searchResult.MaxScore,
		Documents: make([]Document, 0, searchResult.Total),
	}

	for _, hit := range searchResult.Hits {
		doc := Document{
			ID:     hit.ID,
			Score:  hit.Score,
			Fields: make(map[string]string, len(hit.Fields)),
		}

		if query.getHighlight() {
			for key, vals := range hit.Fragments {
				doc.Fields[key] = strings.Join(vals, "\n")
			}
		}

		for key, vv := range hit.Fields {
			if _, ok = doc.Fields[key]; ok {
				continue
			}

			switch val := vv.(type) {
			case string:
				doc.Fields[key] = val
			default:
				continue
			}
		}

		if val, ok := doc.Fields[fieldCreatedAt]; ok {
			if createdAt, err := time.Parse(time.RFC3339, val); err == nil {
				doc.CreatedAt = createdAt
				delete(doc.Fields, fieldCreatedAt)
			}
		}

		result.Documents = append(result.Documents, doc)
	}

	return result, nil
}
