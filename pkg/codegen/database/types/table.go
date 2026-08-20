package types

type TableName interface {
	TableName() string
}

type TableModel struct {
	Name    string
	Columns []string
	Fields  map[string]int
	Attr    []map[string]string
}

const (
	AttrKeyIndexPK     = "pk"
	AttrKeyIndexFK     = "fk"
	AttrKeyIndexUnique = "unq"
	AttrKeyOnCreate    = "on_create"
	AttrKeyOnUpdate    = "on_update"
	AttrKeyOnDelete    = "on_delete"
)

const (
	AttrValAutoIncrement = "auto_increment"
)
