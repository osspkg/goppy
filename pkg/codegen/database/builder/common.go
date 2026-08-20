package builder

import "go.osspkg.com/goppy/v3/pkg/codegen/database/types"

type Query struct {
	Meta *types.TableModel

	Distinct   bool
	Where      []WhereClause
	GroupBy    []string
	OrderBy    []OrderByClause
	LimitVal   int
	OffsetVal  int
	Returning  []string
	Values     [][]any
	OnConflict *ConflictClause

	args     []any
	argCount int
}

type WhereClause struct {
	Column    string
	Operation string
	Value     any
	RawSQL    string
}

type OrderByClause struct {
	Columns []string
	IsDesc  bool
}

type ConflictClause struct {
	Columns []string
	Updates []string
}
