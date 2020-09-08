package solr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uol/solr"
)

func TestSearchSolrLib(t *testing.T) {

	keyset := randomKeyset()
	metric := randomMetric()

	createCollection(t, keyset)

	tags := make(map[string]string)
	tags["host"] = "host"

	_, err := makeDocs(metric, keyset, "", tags, 30, false)
	if !assert.NoError(t, err) {
		return
	}

	searchByMetric(t, keyset, metric)

	searchByTag(t, keyset)

	tags2 := make(map[string]string)
	tags2["filter"] = "tags"

	_, err = makeDocs(metric, keyset, "", tags2, 10, false)
	if !assert.NoError(t, err) {
		return
	}

	tags3 := make(map[string]string)
	tags3["service"] = "solr"

	expected, err := makeDocs(metric, keyset, "", tags3, 10, false)
	if !assert.NoError(t, err) {
		return
	}

	actual, err := searchByTags(t, keyset, "service", "solr", 0, 10)
	if !assert.NoError(t, err) {
		return
	}

	testDocumentRaw(t, expected, actual)

	searchNoDocument(t, keyset)

}

func searchByMetric(t *testing.T, keyset, metric string) {

	fc := make(map[string]string)
	fc["facet.field"] = "metric"

	q := "{!parent which=\"(parent_doc:true AND type:meta AND metric:" + fmt.Sprintf("%s", metric) + ")\"}"
	fl := "*,[child parentFilter=parent_doc:true limit=10000]"
	searchParams := &solr.SearchParams{
		Q:      q,
		FL:     fl,
		Facets: fc,
		Rows:   10,
	}
	libres, err := defaultInstance.Search(searchParams, keyset)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.Equal(t, int64(0), libres.Status) {
		return
	}

	if !assert.NotNil(t, libres.Docs) {
		return
	}

	docs := libres.Docs.([]solr.DocumentRaw)

	if !assert.Equal(t, 10, len(docs)) {
		return
	}

	for i := 0; i < len(docs); i++ {

		if !assert.Equal(t, fmt.Sprintf("host%d", i+1), docs[i]["tag_value"]) {
			return
		}

	}

	for j := 0; j < len(libres.Facets); j++ {

		if !assert.Equal(t, metric, libres.Facets[j].List[j].Name) {
			return
		}
	}

}

func searchByTag(t *testing.T, keyset string) {

	searchParams := &solr.SearchParams{
		Q:    "tag_value:host5",
		Rows: 1,
	}

	res, err := defaultInstance.Search(searchParams, keyset)
	if !assert.NoError(t, err) {
		return
	}

	docs := res.Docs.([]solr.DocumentRaw)
	if assert.Equal(t, "host5", docs[0]["tag_value"]) {
		return
	}

}

func searchByTags(t *testing.T, keyset, key, value string, start, rows int) ([]solr.DocumentRaw, error) {

	searchParams := &solr.SearchParams{
		Q:     "tag_key:" + key + " AND tag_value:/" + value + ".*/",
		Start: start,
		Rows:  rows,
	}

	res, err := defaultInstance.Search(searchParams, keyset)
	if !assert.NoError(t, err) {
		return nil, err
	}

	return res.Docs.([]solr.DocumentRaw), nil

}

func searchNoDocument(t *testing.T, keyset string) {

	searchParams := &solr.SearchParams{
		Q: "id:10000000",
	}

	res, err := defaultInstance.Search(searchParams, keyset)
	if !assert.NoError(t, err) {
		return
	}

	docs := res.Docs.([]solr.DocumentRaw)

	if !assert.Equal(t, 0, len(docs)) {
		return
	}

}
