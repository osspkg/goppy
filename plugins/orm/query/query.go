package query

type Query interface {
	build(prefix string, index *int, args *[]any) string
}

type SQL interface {
	Query() string
	Args() []any
}
