/*
 *  Copyright (c) 2022-2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package search

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	bleve "github.com/blevesearch/bleve/v2"
	customAnalyzer "github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	"github.com/blevesearch/bleve/v2/analysis/token/camelcase"
	"github.com/blevesearch/bleve/v2/analysis/token/lowercase"
	"github.com/blevesearch/bleve/v2/analysis/token/unicodenorm"
	"github.com/blevesearch/bleve/v2/analysis/tokenizer/unicode"
	"github.com/blevesearch/bleve/v2/mapping"
	"go.osspkg.com/goppy/errors"
	"go.osspkg.com/goppy/iofile"
)

type (
	Searcher interface {
		Generate() error
		Open() error
		Close() error
		Add(name, id string, data interface{}) error
		Delete(name, id string) error
		Search(ctx context.Context, name, query string, highlight bool, result interface{}) error
	}

	serviceSearch struct {
		conf    ConfigItem
		indexes map[string]*indexItem
	}

	indexItem struct {
		Index      bleve.Index
		Fields     []string
		FieldsType map[string]string
	}
)

func NewSearch(c ConfigItem) Searcher {
	return &serviceSearch{
		conf:    c,
		indexes: make(map[string]*indexItem, len(c.Indexes)),
	}
}

func (v *serviceSearch) validate() error {
	for _, index := range v.conf.Indexes {
		if !isValidIndexName(index.Name) {
			return fmt.Errorf("invalid index name `%s`, use chars: 0-9 a-z _", index.Name)
		}
		if len(index.Fields) == 0 {
			return fmt.Errorf("index field is empty for `%s`", index.Name)
		}
		for _, field := range index.Fields {
			if !isUpperCamelCase(field.Name) {
				return fmt.Errorf("invalid index field name `%s`, use UpperCamelCase with char: A-Z a-z", field.Name)
			}
			switch field.Type {
			case FieldText, FieldDate:
				continue
			default:
				return fmt.Errorf("invalid index field type `%s`, use text or date", field.Name)
			}
		}
	}
	return nil
}

func (v *serviceSearch) Generate() error {
	if err := v.validate(); err != nil {
		return err
	}

	for _, conf := range v.conf.Indexes {
		folderPath := v.conf.Folder + "/" + conf.Name
		if iofile.Exist(folderPath + indexFilename) {
			continue
		}
		docMapping := bleve.NewDocumentMapping()
		for _, field := range conf.Fields {
			var fieldMapping *mapping.FieldMapping
			switch field.Type {
			case FieldText:
				fieldMapping = bleve.NewTextFieldMapping()
			case FieldDate:
				fieldMapping = bleve.NewDateTimeFieldMapping()
				fieldMapping.Index = false
				fieldMapping.IncludeTermVectors = false
			default:
				return fmt.Errorf("invalid field type `%s`", field.Name)
			}
			fieldMapping.IncludeInAll = false
			docMapping.AddFieldMappingsAt(field.Name, fieldMapping)
		}
		indexMapping := bleve.NewIndexMapping()
		if err := indexMapping.AddCustomTokenFilter(customFilterName, map[string]any{
			"type": unicodenorm.Name,
			"form": unicodenorm.NFC,
		}); err != nil {
			return err
		}
		if err := indexMapping.AddCustomAnalyzer(customAnalyzerName, map[string]any{
			"type":          customAnalyzer.Name,
			"char_filters":  []string{},
			"tokenizer":     unicode.Name,
			"token_filters": []string{customFilterName, camelcase.Name, lowercase.Name},
		}); err != nil {
			return err
		}
		indexMapping.AddDocumentMapping(customAnalyzerName, docMapping)
		indexMapping.AddDocumentMapping("_all", bleve.NewDocumentDisabledMapping())

		if err := os.MkdirAll(folderPath, 0777); err != nil {
			return err
		}
		bleveindex, err := bleve.New(folderPath, indexMapping)
		if err != nil {
			return err
		}
		if err = bleveindex.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (v *serviceSearch) Open() error {
	if err := v.validate(); err != nil {
		return err
	}

	for _, conf := range v.conf.Indexes {
		folderPath := v.conf.Folder + "/" + conf.Name
		if !iofile.Exist(folderPath + indexFilename) {
			return errors.Wrap(
				fmt.Errorf("index not found `%s`", conf.Name),
				v.Close(),
			)
		}
		index, err := bleve.Open(folderPath)
		if err != nil {
			return errors.Wrap(
				fmt.Errorf("index open `%s`: %w", conf.Name, err),
				v.Close(),
			)
		}
		fields := make([]string, 0, len(conf.Fields))
		fieldsType := make(map[string]string, len(conf.Fields))
		for _, field := range conf.Fields {
			fields = append(fields, field.Name)
			fieldsType[field.Name] = field.Type
		}
		v.indexes[conf.Name] = &indexItem{
			Index:      index,
			Fields:     fields,
			FieldsType: fieldsType,
		}
	}

	return nil
}

func (v *serviceSearch) Close() error {
	var err error
	for _, conf := range v.conf.Indexes {
		indx, ok := v.indexes[conf.Name]
		if !ok {
			continue
		}
		delete(v.indexes, conf.Name)
		err = errors.Wrap(indx.Index.Close())
	}
	if err != nil {
		return err
	}
	return nil
}

func (v *serviceSearch) Add(name, id string, data interface{}) error {
	indx, ok := v.indexes[name]
	if !ok {
		return fmt.Errorf("index not found: %s", name)
	}
	return indx.Index.Index(id, data)
}

func (v *serviceSearch) Delete(name, id string) error {
	indx, ok := v.indexes[name]
	if !ok {
		return fmt.Errorf("index not found: %s", name)
	}
	return indx.Index.Delete(id)
}

// nolint: gocyclo
func (v *serviceSearch) Search(ctx context.Context, name, query string, highlight bool, result interface{}) error {
	indx, ok := v.indexes[name]
	if !ok {
		return fmt.Errorf("index not found: %s", name)
	}
	rv := reflect.ValueOf(result)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("result parameter is non-pointer %T", result)
	}
	if rv.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("result parameter is not pointer of slice")
	}
	re := rv.Elem().Type().Elem()
	if re.Kind() != reflect.Struct {
		return fmt.Errorf("result parameter must have struct of slice")
	}
	request := bleve.NewSearchRequestOptions(bleve.NewQueryStringQuery(query), rv.Elem().Cap(), 0, false)
	request.Fields = append(request.Fields, indx.Fields...)
	request.IncludeLocations = true
	if highlight {
		request.Highlight = bleve.NewHighlight()
	}
	searchResult, err := indx.Index.SearchInContext(ctx, request)
	if err != nil {
		return err
	}
	if searchResult.Total == 0 {
		return nil
	}
	for _, hit := range searchResult.Hits {
		obj := reflect.New(re)
		obj.Elem().FieldByName(structScoreField).Set(reflect.ValueOf(hit.Score))
		for key, value := range hit.Fields {
			t, ok := indx.FieldsType[key]
			if !ok {
				continue
			}
			switch t {
			case FieldText:
				if highlight {
					if vv, ok := hit.Fragments[key]; ok {
						value = strings.Join(vv, "")
					}
				}
				obj.Elem().FieldByName(key).Set(reflect.ValueOf(value))
			case FieldDate:
				tv, err := time.Parse(time.RFC3339, value.(string))
				if err != nil {
					return err
				}
				obj.Elem().FieldByName(key).Set(reflect.ValueOf(tv))
			default:
			}

		}
		rv.Elem().Set(reflect.Append(rv.Elem(), obj.Elem()))
	}
	return nil
}
