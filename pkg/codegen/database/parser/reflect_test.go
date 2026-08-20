package parser_test

import (
	"fmt"
	"testing"
	"time"

	"go.osspkg.com/casecheck"
	"go.osspkg.com/goppy/v3/pkg/codegen/database/parser"
)

type MockModel struct {
	ID        uint64    `db:"id;pk=auto_increment"`
	Key       string    `db:"key;unique;on_create=uuid"`
	Value     []byte    `db:"value"`
	CreatedAt time.Time `db:"created_at;on_create=now"`
	UpdatedAt time.Time `db:"updated_at;on_create=now;on_update=now"`
}

func (MockModel) TableName() string {
	return "mock_table"
}

type MockModel2 struct {
	Key   string `db:"key;unique;on_create=uuid"`
	Value []byte `db:"value"`
}

func (MockModel2) TableName() string {
	return "mock_table"
}

func TestUnit_Parse(t *testing.T) {
	info, err := parser.Model(&MockModel{})
	casecheck.NoError(t, err)
	fmt.Println("TableName:", info.Name)
	fmt.Println("Columns:", info.Columns)
	fmt.Println("FieldIndex:", info.FieldIndex)
	fmt.Println("Attr:", info.Attr)

	info, err = parser.Model(MockModel2{})
	casecheck.NoError(t, err)
	fmt.Println("TableName:", info.Name)
	fmt.Println("Columns:", info.Columns)
	fmt.Println("FieldIndex:", info.FieldIndex)
	fmt.Println("Attr:", info.Attr)
}

func Benchmark_Parse(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := parser.Model(MockModel{}); err != nil {
				b.Fatal(err)
			}
		}
	})
}
