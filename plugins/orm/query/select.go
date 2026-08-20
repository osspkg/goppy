package query

import (
	"go.osspkg.com/goppy/v3/pkg/codegen/database/parser"
	"go.osspkg.com/goppy/v3/pkg/codegen/database/types"
)

type SelectQuery interface {
	Distinct() SelectQuery
}

type selectObject struct {
	__meta *types.TableModel

	columns     []string
	distinct    bool
	whereParts  []whereObject
	orderBy     []string
	groupBy     []string
	havingParts []whereObject
	limit       int
	offset      int
	argNum      int
}

func Select[T types.TableName](args ...Query) (SQL, error) {
	meta, err := parser.Generic[T]()
	if err != nil {
		return nil, err
	}
	return &selectObject{__meta: meta}
}
