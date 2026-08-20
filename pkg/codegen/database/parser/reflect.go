package parser

import (
	"fmt"
	"reflect"
	"strings"

	"go.osspkg.com/goppy/v3/pkg/codegen/database/types"
	"go.osspkg.com/ioutils/cache"
)

const (
	tagMain = "db"
)

var tableCache = cache.New[reflect.Type, *types.TableModel]()

func Model(model types.TableName) (*types.TableModel, error) {
	if model == nil {
		return nil, fmt.Errorf("database model %T is nil", model)
	}

	ref := reflect.TypeOf(model)
	if ref.Kind() == reflect.Ptr {
		ref = ref.Elem()
	}
	if ref.Kind() != reflect.Struct {
		return nil, fmt.Errorf("database model %T must be a struct", model)
	}

	if info, ok := tableCache.Get(ref); ok {
		return info, nil
	}

	numField := ref.NumField()
	info := &types.TableModel{
		Name:    model.TableName(),
		Columns: make([]string, 0, numField),
		Fields:  make(map[string]int, numField),
		Attr:    make([]map[string]string, 0, numField),
	}

	for i := 0; i < numField; i++ {
		field := ref.Field(i)

		if !field.IsExported() {
			continue
		}

		tags := field.Tag.Get(tagMain)
		if tags == "-" || tags == "" {
			continue
		}

		list := strings.Split(tags, ";")
		for idx, tag := range list {
			if idx == 0 {
				if _, ok := info.Fields[tag]; ok {
					return nil, fmt.Errorf("database model %T contains duplicate column: %s", model, tag)
				}
				info.Columns = append(info.Columns, tag)
				info.Fields[tag] = i
				info.Attr = append(info.Attr, make(map[string]string, len(list)))
				continue
			}

			attr := strings.Split(tag, "=")
			switch len(attr) {
			case 0:
			case 1:
				info.Attr[i][attr[0]] = ""
			case 2:
				info.Attr[i][attr[0]] = attr[1]
			default:
				return nil, fmt.Errorf(
					"database model %T column %s must have a single '=' attribute",
					model, info.Columns[i],
				)
			}
		}
	}

	if len(info.Name) == 0 || len(info.Columns) == 0 {
		return nil, fmt.Errorf("database model %T must have a single column", model)
	}

	tableCache.Set(ref, info)

	return info, nil
}
