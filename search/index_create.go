/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package search

import (
	"crypto/sha1"
	"fmt"
	"os"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/char/asciifolding"
	"github.com/blevesearch/bleve/v2/analysis/char/html"
	"github.com/blevesearch/bleve/v2/analysis/char/zerowidthnonjoiner"
	"github.com/blevesearch/bleve/v2/analysis/token/camelcase"
	"github.com/blevesearch/bleve/v2/analysis/token/lowercase"
	"github.com/blevesearch/bleve/v2/analysis/token/unicodenorm"
	"github.com/blevesearch/bleve/v2/analysis/token/unique"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/unicode"
	"go.osspkg.com/do"
	"go.osspkg.com/ioutils/fs"
)

const indexRootPath = "/store/root.bolt"

func (v *service) CreateIndex(name string, fields []string) error {
	if inx, ok := v.list.Get(name); ok {
		inxFields, err := inx.Fields()
		if err != nil {
			return fmt.Errorf("get fields exist index: %w", err)
		}

		if diff := do.Diff(inxFields, fields); len(diff) != 0 {
			return fmt.Errorf(
				"fields other than the current index have been set, has fields: %s",
				strings.Join(inxFields, ","),
			)
		}

		return nil
	}

	if err := v.conf.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	if len(fields) == 0 {
		return fmt.Errorf("must specify at least one field")
	}

	folderPath := fmt.Sprintf("%s/%x", v.conf.Folder, sha1.Sum([]byte(name))) //nolint:gosec

	if !fs.FileExist(folderPath + indexRootPath) {
		if err := os.MkdirAll(folderPath, 0744); err != nil {
			return fmt.Errorf("create index directory: %w", err)
		}

		docMapping := bleve.NewDocumentMapping()
		for _, field := range fields {
			fieldMapping := bleve.NewTextFieldMapping()
			fieldMapping.Index = true
			fieldMapping.Store = true
			fieldMapping.IncludeInAll = false
			fieldMapping.IncludeTermVectors = true
			//fieldMapping.Analyzer = "edgeGramAnalyzer"
			docMapping.AddFieldMappingsAt(field, fieldMapping)
		}

		fieldMapping := bleve.NewDateTimeFieldMapping()
		fieldMapping.Index = false
		fieldMapping.IncludeTermVectors = false
		fieldMapping.IncludeInAll = false
		docMapping.AddFieldMappingsAt(fieldCreatedAt, fieldMapping)

		indexMapping := bleve.NewIndexMapping()

		if err := indexMapping.AddCustomTokenFilter(customFilterName, map[string]any{
			"type": unicodenorm.Name,
			"form": unicodenorm.NFC,
		}); err != nil {
			return fmt.Errorf("add custom token filter: %w", err)
		}

		if err := indexMapping.AddCustomAnalyzer(customAnalyzerName, map[string]any{
			"type":          custom.Name,
			"char_filters":  []string{asciifolding.Name, zerowidthnonjoiner.Name, html.Name},
			"tokenizer":     unicode.Name,
			"token_filters": []string{customFilterName, camelcase.Name, lowercase.Name, unique.Name},
		}); err != nil {
			return fmt.Errorf("add custom analyzer: %w", err)
		}

		indexMapping.AddDocumentMapping(customAnalyzerName, docMapping)
		indexMapping.AddDocumentMapping("_all", bleve.NewDocumentDisabledMapping())

		index, err := bleve.New(folderPath, indexMapping)
		if err != nil {
			return fmt.Errorf("create index: %w", err)
		}
		index.SetName(name)
		if err = index.Close(); err != nil {
			return fmt.Errorf("close index: %w", err)
		}
	}

	index, err := bleve.Open(folderPath)
	if err != nil {
		return fmt.Errorf("open index: %w", err)
	}

	v.list.Set(name, index)

	return nil
}
