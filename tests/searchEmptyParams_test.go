package solr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/uol/solr"
)

func TestEmptyParams(t *testing.T) {

	keyset := randomKeyset()
	metric := randomMetric()

	createCollection(t, keyset)

	tags := make(map[string]string)
	tags["host"] = "host"
	makeDocs(metric, keyset, "", tags, 30, false)

	searchParams := &solr.SearchParams{
		Q: "*:*",
	}

	res, err := defaultInstance.Search(searchParams, keyset)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.Equal(t, int64(0), res.Status) {
		return
	}

	if !assert.Equal(t, int64(30), res.NumFound) {
		return
	}
	docs := res.Docs.(([]solr.DocumentRaw))
	if !assert.Equal(t, int(0), len(docs)) {
		return
	}

	searchParams = &solr.SearchParams{
		Q: "tag_value:host1899",
	}

	res, err = defaultInstance.Search(searchParams, keyset)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.Equal(t, int64(0), res.Status) {
		return
	}

	if !assert.Equal(t, int64(0), res.NumFound) {
		return
	}

	if !assert.Empty(t, res.Docs) {
		return
	}

	searchParams = &solr.SearchParams{
		Q: "",
	}

	res, err = defaultInstance.Search(searchParams, keyset)
	if !assert.NoError(t, err) {
		return
	}

	if !assert.Equal(t, int64(0), res.Status) {
		return
	}

	if !assert.Equal(t, int64(0), res.NumFound) {
		return
	}

	if !assert.Empty(t, res.Docs) {
		return
	}

}
