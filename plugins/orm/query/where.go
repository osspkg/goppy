package query

import "fmt"

type whereObject struct {
	col  string
	cond string
	arg  any
	sep  string
}

func (w whereObject) build(prefix string, index *int, args *[]any) string {
	*index++
	*args = append(*args, w.arg)
	return fmt.Sprintf("%s %s\"%s\"%s$%d", w.sep, prefix, w.col, w.cond, *index)
}

func Where(field, op string, val any) Query {
	return whereObject{
		col:  field,
		cond: op,
		arg:  val,
		sep:  "AND",
	}
}

func OrWhere(field, op string, val any) Query {
	return whereObject{
		col:  field,
		cond: op,
		arg:  val,
		sep:  "OR",
	}
}

func AndWhere(field, op string, val any) Query {

}
