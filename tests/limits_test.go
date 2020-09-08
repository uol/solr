package solr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uol/solr"
)

func TestLimits(t *testing.T) {

	keyset := randomKeyset()
	metric := randomMetric()

	createCollection(t, keyset)

	tags := make(map[string]string)
	tags["host"] = "host"

	_, err := makeDocs(metric, keyset, "", tags, 30, false)
	if !assert.NoError(t, err) {
		return
	}

	//test start 0 and rows 0 expected is 0 rows
	testStartRows(t, keyset, 0, 0, 0)

	//test start 10 and rows 10 expected is 10 rows, rows total 30
	testStartRows(t, keyset, 10, 10, 10)

	//test start 25 rows and rows 5 expected is 5, 5 rows remaining
	testStartRows(t, keyset, 25, 5, 5)

	//test start 35 rows 10 expected 0, exceeded the limit
	testStartRows(t, keyset, 35, 10, 0)

}

func testStartRows(t *testing.T, instanceName string, start, rows, expected int) {

	searchParams := &solr.SearchParams{
		Q:     "*:*",
		Start: start,
		Rows:  rows,
	}
	res, err := defaultInstance.Search(searchParams, instanceName)
	if !assert.NoError(t, err) {
		t.Log(err)
		t.Fail()
	}

	docs := res.Docs.([]solr.DocumentRaw)

	if !assert.True(t, expected == len(docs)) {
		t.Log("greater than expected amount of metadata")
		t.Fail()
	}

}
