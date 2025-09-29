/*
 *  Copyright (c) 2022-2025 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a BSD 3-Clause license that can be found in the LICENSE file.
 */

package search_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"go.osspkg.com/casecheck"

	"go.osspkg.com/goppy/v2/search"
)

func dumpResult(v *search.Result) {
	fmt.Println("-------------------------------------------------------------------------------------------------")
	for _, item := range v.Documents {
		fmt.Printf("-  ID: %s, Score: %f, CreatedAt: %s\n", item.ID, item.Score, item.CreatedAt.Format(time.RFC3339))
		for key, val := range item.Fields {
			fmt.Printf("   %s: %s\n", key, val)
		}
	}
	fmt.Printf("Found: %d items, time: %v\n", v.Total, v.Took)
	fmt.Println("-------------------------------------------------------------------------------------------------")
}

func TestUnit_NewSearch(t *testing.T) {
	conf := search.Config{
		Folder: "/tmp/TestUnit_NewSearch",
	}

	indexName := "index name"

	defer func() {
		casecheck.NoError(t, os.RemoveAll(conf.Folder))
	}()

	srv := search.New(conf)
	casecheck.NoError(t, srv.CreateIndex(indexName, []string{"line", "data"}))
	defer func() {
		casecheck.NoError(t, srv.Close())
	}()

	data := []string{
		"The Golden Eye of Day",
		"A radiant sphere, a cosmic flame,",
		"The sun awakes, whispering its name.",
		"It climbs the blue, a monarch on high,",
		"Painting the canvas of the eastern sky.",
		"Golden threads unspooling, bright and fast,",
		"Warming the air, shadows quickly past.",
		"A tireless furnace, burning ever bright,",
		"Dispelling darkness with its glorious light.",
		"The dawn unfurls in hues of rose and gold,",
		"A story ancient, endlessly retold.",
		"It kisses mountains, wakes the sleepy seas,",
		"And stirs the chatter in the rustling trees.",
		"A gentle giant, pouring out its power,",
		"Nourishing life in every passing hour.",
		"The <b>solar</b> gaze, intense and strong and pure,",
		"A constant promise, certain to endure.",
		"Through hazy mists or skies completely clear,",
		"Its cheerful presence banishes all fear.",
		"From dewdrops catching its prismatic gleam,",
		"To basking creatures living out a dream.",
		"It rules the hours, dictating work and play,",
		"The central actor of the earthly day.",
		"A fiery engine driving wind and tide,",
		"With nothing on this planet left to hide.",
		"It sets at last, in crimson, violet, deep,",
		"A solemn promise that it won't just sleep.",
		"But travel onward, to return anew,",
		"In twenty-four hours, its daily work to do.",
		"A burning heart, so distant, yet so near,",
		"The sun, our anchor, year after patient year.",
		"A perfect circle, brilliant and immense,",
		"Our source of warmth, our vital recompense.",
	}

	casecheck.NoError(t, srv.AddDocuments(indexName, "orig ", map[string]string{
		"line": "1",
		"data": strings.Join(data, " "),
	}))

	for i, datum := range data {
		casecheck.NoError(t, srv.AddDocuments(indexName, fmt.Sprintf("x%d", i), map[string]string{
			"line": fmt.Sprintf("%d", i),
			"data": datum,
		}))
	}

	//***************************************************************************************************

	result, err := srv.Search(context.TODO(), search.MatchQuery{
		Query:      "burning ever solar",
		IndexName:  indexName,
		ShowFields: []string{"line", "data"},
		Highlight:  false,
		Limit:      10,
	})
	casecheck.NoError(t, err)
	dumpResult(result)
	casecheck.Equal(t, uint64(4), result.Total)

	//***************************************************************************************************

	casecheck.NoError(t, srv.DeleteDocuments(indexName, "x1"))

	result, err = srv.Search(context.TODO(), search.MatchAllQuery{
		IndexName: indexName,
		Limit:     10,
	})
	casecheck.NoError(t, err)
	dumpResult(result)
	casecheck.Equal(t, uint64(33), result.Total)

	//***************************************************************************************************

	result, err = srv.Search(context.TODO(), search.MatchUniversalQuery{
		Query:     "h* worl* solar",
		IndexName: indexName,
		From:      0,
		Highlight: true,
		Limit:     10,
	})
	casecheck.NoError(t, err)
	dumpResult(result)
	casecheck.Equal(t, uint64(10), result.Total)

}
