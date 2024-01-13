package search_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"go.osspkg.com/goppy/search"
	"go.osspkg.com/goppy/xtest"
)

type TestData struct {
	Line      string
	Data      string
	CreatedAt time.Time
}

type TestDataSearch struct {
	Score     float64
	Line      string
	Data      string
	CreatedAt time.Time
}

func TestUnit_NewSearch(t *testing.T) {
	conf := search.ConfigItem{
		Folder: "/tmp/TestUnit_NewSearch",
		Indexes: []search.ConfigIndex{
			{Name: "demo", Fields: []search.ConfigIndexField{
				{Name: "Line", Type: "text"},
				{Name: "Data", Type: "text"},
				{Name: "CreatedAt", Type: "date"},
			}},
		},
	}

	srv := search.NewSearch(conf)

	xtest.NoError(t, srv.Generate())
	xtest.NoError(t, srv.Open())
	defer func() {
		xtest.NoError(t, os.RemoveAll(conf.Folder))
	}()
	xtest.NoError(t, srv.Add("demo", "001", TestData{
		Line:      "1",
		Data:      "Hello world",
		CreatedAt: time.Now(),
	}))
	xtest.NoError(t, srv.Add("demo", "002", TestData{
		Line:      "2",
		Data:      "Happy world",
		CreatedAt: time.Now(),
	}))
	result := make([]TestDataSearch, 0, 10)
	//xtest.NoError(t, srv.Search("demo", "h* worl*", &result))
	xtest.NoError(t, srv.Search(context.TODO(), "demo", "Hello world", true, &result))
	fmt.Println(result)
	xtest.True(t, len(result) == 2)
	xtest.NoError(t, srv.Delete("demo", "002"))
	result = make([]TestDataSearch, 0, 10)
	xtest.NoError(t, srv.Search(context.TODO(), "demo", "h* worl*", true, &result))
	xtest.True(t, len(result) == 1)
	xtest.NoError(t, srv.Close())
}
